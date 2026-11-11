package security

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownloadVerifiedRejectsUntrustedURL(t *testing.T) {
	client := &http.Client{}
	err := DownloadVerified(context.Background(), client,
		"http://example.com/file", strings.Repeat("a", 64),
		filepath.Join(t.TempDir(), "update.bin"), DefaultMaxDownloadBytes)
	if err == nil {
		t.Fatal("expected non-HTTPS URL to be rejected")
	}
}

func TestDownloadVerifiedRequiresDigest(t *testing.T) {
	client := &http.Client{}
	err := DownloadVerified(context.Background(), client,
		"https://github.com/dessant/buster-client/releases/download/v0.3.0/asset",
		"", filepath.Join(t.TempDir(), "update.bin"), DefaultMaxDownloadBytes)
	if err == nil {
		t.Fatal("expected missing digest to be rejected")
	}
}
