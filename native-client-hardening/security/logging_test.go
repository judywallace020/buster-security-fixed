package security

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestNewPrivateLoggerCreatesRegularFile(t *testing.T) {
	oldHome := os.Getenv("XDG_CACHE_HOME")
	cache, err := ioutil.TempDir("", "buster-cache-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cache)
	if err := os.Setenv("XDG_CACHE_HOME", cache); err != nil {
		t.Fatal(err)
	}
	defer os.Setenv("XDG_CACHE_HOME", oldHome)

	logger, file, err := NewPrivateLogger("test.log")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	logger.Print("redacted test message")

	info, err := os.Stat(filepath.Join(cache, "buster-client", "test.log"))
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsRegular() {
		t.Fatal("log target is not a regular file")
	}
}

func TestNewPrivateLoggerRejectsSymlinkTarget(t *testing.T) {
	oldHome := os.Getenv("XDG_CACHE_HOME")
	cache, err := ioutil.TempDir("", "buster-cache-symlink-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cache)
	if err := os.Setenv("XDG_CACHE_HOME", cache); err != nil {
		t.Fatal(err)
	}
	defer os.Setenv("XDG_CACHE_HOME", oldHome)

	appDir := filepath.Join(cache, "buster-client")
	if err := os.MkdirAll(appDir, 0700); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(cache, "outside.log")
	if err := ioutil.WriteFile(outside, []byte("do not append"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(appDir, "linked.log")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if _, _, err := NewPrivateLogger("linked.log"); err == nil {
		t.Fatal("symlink log target must be rejected")
	}
	data, err := ioutil.ReadFile(outside)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "do not append" {
		t.Fatalf("symlink target was modified: %q", data)
	}
}

func TestNewPrivateLoggerRejectsInvalidFilename(t *testing.T) {
	if _, _, err := NewPrivateLogger("../outside.log"); err == nil {
		t.Fatal("path traversal in log filename must be rejected")
	}
}
