package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestValidRequestOriginForMethod(t *testing.T) {
	const host = "127.0.0.1:43210"
	chrome := "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	firefox := "moz-extension://12345678-1234-1234-1234-123456789abc"

	tests := []struct {
		name, method, origin string
		want bool
	}{
		{"top-level GET without Origin", "GET", "", true},
		{"POST without Origin", "POST", "", false},
		{"POST null Origin", "POST", "null", false},
		{"POST trusted Chrome extension", "POST", chrome, true},
		{"POST trusted Firefox extension", "POST", firefox, true},
		{"POST exact local setup origin", "POST", "http://" + host, true},
		{"POST hostile website", "POST", "https://attacker.example", false},
		{"DELETE without Origin", "DELETE", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := validRequestOriginForMethod(tc.method, tc.origin, host); got != tc.want {
				t.Fatalf("validRequestOriginForMethod(%q, %q) = %v, want %v", tc.method, tc.origin, got, tc.want)
			}
		})
	}
}

func TestValidExtensionIdentifier(t *testing.T) {
	if !validExtensionIdentifier("chrome", "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/") {
		t.Fatal("expected valid Chrome extension origin")
	}
	if validExtensionIdentifier("chrome", "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Fatal("Chrome origin without required trailing slash must be rejected")
	}
	if !validExtensionIdentifier("firefox", "buster@example.org") {
		t.Fatal("expected valid Firefox extension identifier")
	}
	if validExtensionIdentifier("firefox", "../../outside") {
		t.Fatal("path traversal must not be a Firefox extension identifier")
	}
}

func TestIsWithinBase(t *testing.T) {
	base, cleanup := makeTempDir(t)
	defer cleanup()
	inside := filepath.Join(base, "app", "client")
	outside := filepath.Join(filepath.Dir(base), "outside")
	if !isWithinBase(inside, base) {
		t.Fatal("expected nested path to be within base")
	}
	if isWithinBase(base, base) {
		t.Fatal("base itself must not be accepted as an install target")
	}
	if isWithinBase(outside, base) {
		t.Fatal("sibling path must not be within base")
	}
	if isWithinBase(filepath.Join(base, "..", "outside"), base) {
		t.Fatal("traversal path must not be within base")
	}
	if isWithinBase("relative/path", base) {
		t.Fatal("relative path must not be accepted")
	}
}

func TestEnsureNoSymlinkComponents(t *testing.T) {
	base, cleanup := makeTempDir(t)
	defer cleanup()
	target := filepath.Join(base, "normal", "child")
	if err := os.MkdirAll(target, 0700); err != nil {
		t.Fatal(err)
	}
	if err := ensureNoSymlinkComponents(target); err != nil {
		t.Fatalf("normal directory path rejected: %v", err)
	}

	link := filepath.Join(base, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink creation unavailable on this platform: %v", err)
	}
	if err := ensureNoSymlinkComponents(filepath.Join(link, "child")); err == nil {
		t.Fatal("symlink component must be rejected")
	}
}

func TestPathExistsRejectsSymlink(t *testing.T) {
	base, cleanup := makeTempDir(t)
	defer cleanup()
	regular := filepath.Join(base, "manifest.json")
	if err := ioutil.WriteFile(regular, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if exists, err := pathExists(regular); err != nil || !exists {
		t.Fatalf("regular file: exists=%v err=%v", exists, err)
	}
	link := filepath.Join(base, "manifest-link.json")
	if err := os.Symlink(regular, link); err != nil {
		t.Skipf("symlink creation unavailable on this platform: %v", err)
	}
	if _, err := pathExists(link); err == nil {
		t.Fatal("manifest symlink must be rejected")
	}
}

func makeTempDir(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := ioutil.TempDir("", "buster-setup-test-")
	if err != nil {
		t.Fatal(err)
	}
	return dir, func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("remove temporary directory: %v", err)
		}
	}
}
