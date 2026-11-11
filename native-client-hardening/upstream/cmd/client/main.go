package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dessant/nativemessaging"

	"buster-client-hardening/security"
	"buster-client/pkg/input"
	"buster-client/utils"
)

var buildVersion string

type message struct {
	MessageID  string `json:"id"`
	APIVersion string `json:"apiVersion"`
	Command    string `json:"command"`
	Data       string `json:"data"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
}

type response struct {
	MessageID  string `json:"id"`
	APIVersion string `json:"apiVersion"`
	Success    bool   `json:"success"`
	Data       string `json:"data"`
}

func installClient(version string) error {
	if runtime.GOOS != "windows" {
		admin, err := utils.UserAdmin()
		if err != nil {
			log.Println(err)
			return errors.New("cannot inspect current user")
		}
		if admin {
			return errors.New("setup must be run without administrative rights")
		}
	}

	goos := runtime.GOOS
	if goos == "darwin" {
		goos = "macos"
	}
	assetName := fmt.Sprintf("buster-client-setup-v%s-%s-%s", version, goos, runtime.GOARCH)
	if goos == "windows" {
		assetName += ".exe"
	}

	execPath, err := os.Executable()
	if err != nil {
		log.Println(err)
		return errors.New("cannot get executable path")
	}
	setupPath := filepath.Join(filepath.Dir(execPath), utils.GetExecName("buster-client-setup"))
	stageToken, err := security.NewSessionToken()
	if err != nil {
		return errors.New("cannot create a unique updater staging name")
	}
	stagedPath := stagedSetupPath(setupPath, runtime.GOOS, stageToken)

	metadataClient := security.NewGitHubMetadataClient()
	assetURL, digest, err := security.ReleaseAssetSHA256(metadataClient, version, assetName)
	if err != nil {
		log.Printf("Updater metadata validation failed: %v", err)
		return errors.New("cannot verify update release metadata")
	}

	downloadCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := security.DownloadVerified(downloadCtx, security.NewReleaseHTTPClient(), assetURL, digest, stagedPath, security.DefaultMaxDownloadBytes); err != nil {
		log.Printf("Verified updater download failed: %v", err)
		return errors.New("cannot download and verify setup")
	}
	defer os.Remove(stagedPath)

	setupOutput, err := exec.Command(stagedPath, "--update").CombinedOutput()
	if err != nil {
		log.Printf("Verified setup execution failed: %v", err)
		return errors.New("cannot run verified setup")
	}

	if output := string(setupOutput); output != "" {
		return errors.New(output)
	}
	return nil
}

func installCleanup() error {
	execPath, err := os.Executable()
	if err != nil {
		log.Println(err)
		return errors.New("cannot get executable path")
	}
	if err := os.Remove(execPath + ".old"); err != nil && !os.IsNotExist(err) {
		log.Printf("Cleanup failed: %v", err)
		return errors.New("cannot remove previous client executable")
	}
	return nil
}

var inputKeyPattern = regexp.MustCompile(`^[A-Za-z0-9_+.-]{1,64}$`)

const maxNativeTextBytes = 4096
const minNativeCoordinate = -32768
const maxNativeCoordinate = 32767

func validateInputMessage(msg *message) error {
	if msg == nil {
		return errors.New("missing native message")
	}
	switch msg.Command {
	case "pressKey", "releaseKey", "tapKey":
		if !inputKeyPattern.MatchString(msg.Data) {
			return errors.New("invalid key identifier")
		}
	case "typeText":
		if len(msg.Data) > maxNativeTextBytes || !utf8.ValidString(msg.Data) || strings.IndexByte(msg.Data, 0) >= 0 {
			return errors.New("text input is invalid or exceeds 4096 bytes")
		}
	case "moveMouse":
		if msg.X < minNativeCoordinate || msg.X > maxNativeCoordinate ||
			msg.Y < minNativeCoordinate || msg.Y > maxNativeCoordinate {
			return errors.New("mouse coordinates are out of bounds")
		}
	case "clickMouse", "ping":
		// These commands carry no user-controlled payload.
	default:
		return errors.New("unknown command")
	}
	return nil
}

func processMessage(msg *message, rsp *response) error {
	if err := validateInputMessage(msg); err != nil {
		return err
	}
	log.Printf("Command: %s", msg.Command)

	switch msg.Command {
	case "pressKey":
		return input.PressKey(msg.Data)
	case "releaseKey":
		return input.ReleaseKey(msg.Data)
	case "tapKey":
		return input.TapKey(msg.Data)
	case "typeText":
		return input.TypeText(msg.Data)
	case "moveMouse":
		return input.MoveMouse(msg.X, msg.Y)
	case "clickMouse":
		return input.ClickMouse()
	case "ping":
		rsp.Data = "pong"
	}
	return nil
}

func main() {
	_, logFile, logErr := security.NewPrivateLogger("buster-client-log.txt")
	if logErr != nil {
		return
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC)
	log.Printf("Starting client (version: %s)", buildVersion)

	go func() {
		<-time.After(12 * time.Minute)
		log.Println("Closing client (forced)")
		os.Exit(0)
	}()

	decoder := nativemessaging.NewNativeJSONDecoder(os.Stdin)
	encoder := nativemessaging.NewNativeJSONEncoder(os.Stdout)

	for {
		log.Println("Receiving message")
		var msg message
		var rsp response
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				log.Println("Closing client")
			} else {
				log.Println(err)
			}
			return
		}

		rsp.MessageID = msg.MessageID
		rsp.APIVersion = buildVersion

		switch msg.Command {
		case "installClient":
			log.Printf("Installing client (version: %s)", msg.Data)
			if err := installClient(msg.Data); err != nil {
				rsp.Data = err.Error()
			} else {
				rsp.Success = true
			}
		case "installCleanup":
			log.Println("Cleaning up after installation")
			if err := installCleanup(); err != nil {
				rsp.Data = err.Error()
			} else {
				rsp.Success = true
			}
		default:
			if msg.APIVersion != buildVersion {
				log.Printf("Unsupported client version (requested: %s, supported: %s)", msg.APIVersion, buildVersion)
				rsp.Data = "unsupported client version"
			} else if err := processMessage(&msg, &rsp); err != nil {
				rsp.Data = err.Error()
			} else {
				rsp.Success = true
			}
		}

		log.Println("Sending response")
		if err := encoder.Encode(rsp); err != nil {
			log.Println(err)
			return
		}
	}
}

func stagedSetupPath(setupPath, goos, token string) string {
	path := setupPath + "." + token + ".verified"
	if goos == "windows" {
		path += ".exe"
	}
	return path
}
