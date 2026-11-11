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
	"runtime"
	"time"

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
	stagedPath := setupPath + ".verified"

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

	errMessage := string(setupOutput)
	if errMessage != "" {
		return errors.New(errMessage)
	}

	return nil
}

func installCleanup() error {
	execPath, err := os.Executable()
	if err != nil {
		log.Println(err)
		return errors.New("cannot get executable path")
	}

	os.Remove(execPath + ".old")

	return nil
}

func processMessage(msg *message, rsp *response) error {
	log.Printf("Command: %s", msg.Command)

	if msg.Command == "pressKey" {
		input.PressKey(msg.Data)
	} else if msg.Command == "releaseKey" {
		input.ReleaseKey(msg.Data)
	} else if msg.Command == "tapKey" {
		log.Printf("Data: %s", msg.Data)
		input.TapKey(msg.Data)
	} else if msg.Command == "typeText" {
		log.Printf("Data: %s", msg.Data)
		input.TypeText(msg.Data)
	} else if msg.Command == "moveMouse" {
		log.Printf("X: %d, Y: %d", msg.X, msg.Y)
		input.MoveMouse(msg.X, msg.Y)
	} else if msg.Command == "clickMouse" {
		input.ClickMouse()
	} else if msg.Command == "ping" {
		rsp.Data = "pong"
	} else {
		return errors.New("unknown command")
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

		if msg.Command == "installClient" {
			log.Printf("Installing client (version: %s)", msg.Data)
			err := installClient(msg.Data)
			if err == nil {
				rsp.Success = true
			} else {
				rsp.Data = err.Error()
			}
		} else if msg.Command == "installCleanup" {
			log.Println("Cleaning up after installation")
			err := installCleanup()
			if err == nil {
				rsp.Success = true
			} else {
				rsp.Data = err.Error()
			}
		} else {
			if msg.APIVersion == buildVersion {
				log.Println("Processing message")
				err := processMessage(&msg, &rsp)
				if err == nil {
					rsp.Success = true
				} else {
					rsp.Data = err.Error()
				}
			} else {
				log.Printf("Unsupported client version (requested: %s, supported: %s)", msg.APIVersion, buildVersion)
			}
		}

		log.Println("Sending response")

		if err := encoder.Encode(rsp); err != nil {
			log.Println(err)
			return
		}
	}
}
