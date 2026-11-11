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
