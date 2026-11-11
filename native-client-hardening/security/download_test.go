package security

import (
	"context"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func tempTestDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "buster-security-test-")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestDownloadVerifiedRejectsUntrustedURL(t *testing.T) {
	client := &http.Client{}
	err := DownloadVerified(context.Background(), client,
		"http://example.com/file", strings.Repeat("a", 64),
		filepath.Join(tempTestDir(t), "update.bin"), DefaultMaxDownloadBytes)
	if err == nil {
		t.Fatal("expected non-HTTPS URL to be rejected")
	}
}

func TestDownloadVerifiedRequiresDigest(t *testing.T) {
	client := &http.Client{}
	err := DownloadVerified(context.Background(), client,
		"https://github.com/dessant/buster-client/releases/download/v0.3.0/asset",
		"", filepath.Join(tempTestDir(t), "update.bin"), DefaultMaxDownloadBytes)
	if err == nil {
		t.Fatal("expected missing digest to be rejected")
	}
}
