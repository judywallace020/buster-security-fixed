package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
	dir := tempTestDir(t)
	defer os.RemoveAll(dir)
	err := DownloadVerified(context.Background(), client,
		"http://example.com/file", strings.Repeat("a", 64),
		filepath.Join(dir, "update.bin"), DefaultMaxDownloadBytes)
	if err == nil {
		t.Fatal("expected non-HTTPS URL to be rejected")
	}
}

func TestDownloadVerifiedRequiresDigest(t *testing.T) {
	client := &http.Client{}
	dir := tempTestDir(t)
	defer os.RemoveAll(dir)
	err := DownloadVerified(context.Background(), client,
		"https://github.com/dessant/buster-client/releases/download/v0.3.0/asset",
		"", filepath.Join(dir, "update.bin"), DefaultMaxDownloadBytes)
	if err == nil {
		t.Fatal("expected missing digest to be rejected")
	}
}

func downloadTestClient(status int, payload string) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Status:     http.StatusText(status),
			Header:     make(http.Header),
			Body:       testReadCloser{strings.NewReader(payload)},
			ContentLength: int64(len(payload)),
			Request:    r,
		}, nil
	})}
}

func TestDownloadVerifiedWritesOnlyMatchingDigest(t *testing.T) {
	dir := tempTestDir(t)
	defer os.RemoveAll(dir)
	payload := "verified updater payload"
	digest := sha256.Sum256([]byte(payload))
	destination := filepath.Join(dir, "verified.bin")

	err := DownloadVerified(context.Background(), downloadTestClient(http.StatusOK, payload),
		"https://github.com/dessant/buster-client/releases/download/v1.2.3/setup.bin",
		hex.EncodeToString(digest[:]), destination, 1024)
	if err != nil {
		t.Fatalf("valid download rejected: %v", err)
	}
	got, err := ioutil.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != payload {
		t.Fatalf("downloaded payload = %q, want %q", got, payload)
	}
}

func TestDownloadVerifiedRejectsTamperedPayload(t *testing.T) {
	dir := tempTestDir(t)
	defer os.RemoveAll(dir)
	trusted := sha256.Sum256([]byte("trusted payload"))
	destination := filepath.Join(dir, "tampered.bin")

	err := DownloadVerified(context.Background(), downloadTestClient(http.StatusOK, "tampered payload"),
		"https://github.com/dessant/buster-client/releases/download/v1.2.3/setup.bin",
		hex.EncodeToString(trusted[:]), destination, 1024)
	if err == nil {
		t.Fatal("tampered payload must fail digest verification")
	}
	if _, statErr := os.Lstat(destination); !os.IsNotExist(statErr) {
		t.Fatalf("unverified destination should not exist, stat error=%v", statErr)
	}
}

func TestDownloadVerifiedRejectsOversizedAndNonSuccessResponses(t *testing.T) {
	dir := tempTestDir(t)
	defer os.RemoveAll(dir)
	digest := sha256.Sum256([]byte("payload"))
	url := "https://github.com/dessant/buster-client/releases/download/v1.2.3/setup.bin"

	if err := DownloadVerified(context.Background(), downloadTestClient(http.StatusOK, "payload"),
		url, hex.EncodeToString(digest[:]), filepath.Join(dir, "large.bin"), 3); err == nil {
		t.Fatal("oversized payload must be rejected")
	}
	if err := DownloadVerified(context.Background(), downloadTestClient(http.StatusNotFound, "payload"),
		url, hex.EncodeToString(digest[:]), filepath.Join(dir, "status.bin"), 1024); err == nil {
		t.Fatal("non-200 response must be rejected")
	}
}

type testReadCloser struct {
	io.Reader
}

func (testReadCloser) Close() error { return nil }
