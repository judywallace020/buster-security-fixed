// Package security contains hardened helpers for the Buster native client.
// The expected SHA-256 must come from trusted release metadata (for example,
// the GitHub Releases API over verified HTTPS), never from the asset itself.
package security

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const DefaultMaxDownloadBytes int64 = 100 << 20

func DownloadVerified(ctx context.Context, client *http.Client, sourceURL, expectedSHA256, destination string, maxBytes int64) error {
	if client == nil {
		return errors.New("HTTP client is required")
	}
	if maxBytes <= 0 || maxBytes > DefaultMaxDownloadBytes {
		return errors.New("invalid download size limit")
	}
	if len(expectedSHA256) != 64 {
		return errors.New("a SHA-256 digest from trusted release metadata is required")
	}
	expected, err := hex.DecodeString(expectedSHA256)
	if err != nil || len(expected) != sha256.Size {
		return errors.New("invalid SHA-256 digest")
	}

	u, err := url.Parse(sourceURL)
	if err != nil || u.Scheme != "https" || u.Hostname() != "github.com" || u.User != nil {
		return errors.New("download URL must be an HTTPS GitHub release URL")
	}
	if !strings.HasPrefix(u.Path, "/dessant/buster-client/releases/download/") {
		return errors.New("unexpected release download path")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected download status: %s", resp.Status)
	}
	if resp.ContentLength > maxBytes {
		return errors.New("download exceeds size limit")
	}

	dir := filepath.Dir(destination)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".buster-update-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	h := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(tmp, h), io.LimitReader(resp.Body, maxBytes+1))
	if copyErr != nil {
		tmp.Close()
		return copyErr
	}
	if written > maxBytes {
		tmp.Close()
		return errors.New("download exceeds size limit")
	}
	actual := h.Sum(nil)
	if subtle.ConstantTimeCompare(actual, expected) != 1 {
		tmp.Close()
		return errors.New("download SHA-256 verification failed")
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0700); err != nil {
		return err
	}
	if err := os.Rename(tmpName, destination); err != nil {
		return err
	}
	return nil
}

func NewReleaseHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 2 * time.Minute,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many redirects")
			}
			if req.URL.Scheme != "https" {
				return errors.New("refusing non-HTTPS redirect")
			}
			host := strings.ToLower(req.URL.Hostname())
			switch host {
			case "github.com", "objects.githubusercontent.com", "release-assets.githubusercontent.com":
				return nil
			default:
				return fmt.Errorf("refusing redirect to untrusted host %q", host)
			}
		},
	}
}
