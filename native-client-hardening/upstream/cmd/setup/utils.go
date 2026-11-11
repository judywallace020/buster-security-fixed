package main

import (
	"errors"
	"os"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"regexp"
	"strings"

	"buster-client-hardening/security"
)

func stringInSlice(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func pathExists(path string) (bool, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return false, errors.New("manifest path must be a regular non-symlink file")
	}
	return true, nil
}

// ensureNoSymlinkComponents rejects symlinks in every existing component of
// an installation directory. Missing suffixes are safe to create after the
// existing prefix has been checked.
func ensureNoSymlinkComponents(path string) error {
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return err
	}
	volume := filepath.VolumeName(abs)
	current := volume + string(os.PathSeparator)
	remainder := strings.TrimPrefix(abs, current)
	parts := strings.Split(remainder, string(os.PathSeparator))
	for _, part := range parts {
		if part == "" {
			continue
		}
		current = filepath.Join(current, part)
		info, statErr := os.Lstat(current)
		if os.IsNotExist(statErr) {
			return nil
		}
		if statErr != nil {
			return statErr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("installation path contains a symbolic link")
		}
		if !info.IsDir() {
			return errors.New("installation path component is not a directory")
		}
	}
	return nil
}

func newToken() (string, error) {
	return security.NewSessionToken()
}

func restoreAsset(path, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path, data, info.Mode()); err != nil {
		return err
	}
	if err := os.Chtimes(path, info.ModTime(), info.ModTime()); err != nil {
		return err
	}

	return nil
}

var chromeExtensionOriginPattern = regexp.MustCompile(`^chrome-extension://[a-p]{32}/$`)
var firefoxExtensionIDPattern = regexp.MustCompile(`^[A-Za-z0-9._@{}+-]{1,128}$`)

func validBrowserTarget(browser, targetEnv string) bool {
	validBrowser := map[string]bool{
		"chrome": true, "edge": true, "firefox": true, "opera": true,
		"chromium": true, "microsoft edge": true, "waterfox": true,
		"icedragon": true, "cliqz": true,
	}
	validTarget := map[string]bool{
		"chrome": true, "firefox": true, "chromium": true,
		"microsoft edge": true, "edge": true, "opera": true,
	}
	return validBrowser[browser] && validTarget[targetEnv]
}

func validExtensionIdentifier(targetEnv, extension string) bool {
	if targetEnv == "firefox" {
		return firefoxExtensionIDPattern.MatchString(extension)
	}
	return chromeExtensionOriginPattern.MatchString(extension)
}

func sameExpectedPath(supplied, expected string) bool {
	if supplied == "" || expected == "" {
		return supplied == expected
	}
	suppliedAbs, err := filepath.Abs(filepath.Clean(supplied))
	if err != nil {
		return false
	}
	expectedAbs, err := filepath.Abs(filepath.Clean(expected))
	if err != nil {
		return false
	}
	if runtime.GOOS == "windows" {
		return strings.EqualFold(suppliedAbs, expectedAbs)
	}
	return suppliedAbs == expectedAbs
}
