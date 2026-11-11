package security

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"encoding/hex"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const releaseMetadataURL = "https://api.github.com/repos/dessant/buster-client/releases/tags/v"

var releaseVersionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+([-+][0-9A-Za-z.-]+)?$`)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		Digest             string `json:"digest"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// ReleaseAssetSHA256 obtains the SHA-256 published in GitHub release metadata.
// It fails closed when GitHub does not publish a digest for the requested asset.
func ReleaseAssetSHA256(client *http.Client, version, assetName string) (string, string, error) {
	if client == nil {
		return "", "", errors.New("HTTP client is required")
	}
	if !releaseVersionPattern.MatchString(version) {
		return "", "", errors.New("invalid release version")
	}
	if assetName == "" || strings.ContainsAny(assetName, "/\\") {
		return "", "", errors.New("invalid release asset name")
	}

	req, err := http.NewRequest(http.MethodGet, releaseMetadataURL+url.PathEscape(version), nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "buster-client-security-updater")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub release metadata returned %s", resp.Status)
	}

	const maxMetadataBytes = 2 << 20
	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, maxMetadataBytes+1))
	if err != nil {
		return "", "", err
	}
	if int64(len(body)) > maxMetadataBytes {
		return "", "", errors.New("GitHub release metadata exceeds size limit")
	}

	var release githubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", "", err
	}
	if release.TagName != "v"+version {
		return "", "", errors.New("GitHub release tag does not match requested version")
	}

	for _, asset := range release.Assets {
		if asset.Name != assetName {
			continue
		}
		if !strings.HasPrefix(asset.Digest, "sha256:") {
			return "", "", errors.New("release asset has no trusted SHA-256 digest")
		}
		digest := strings.TrimPrefix(asset.Digest, "sha256:")
		if len(digest) != 64 {
			return "", "", errors.New("release asset SHA-256 digest has invalid length")
		}
		decoded, err := hex.DecodeString(digest)
		if err != nil || len(decoded) != 32 {
			return "", "", errors.New("release asset SHA-256 digest is invalid")
		}
		u, err := url.Parse(asset.BrowserDownloadURL)
		if err != nil || u.Scheme != "https" || u.Hostname() != "github.com" || u.Port() != "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			return "", "", errors.New("release asset URL is not a trusted GitHub URL")
		}
		if !strings.HasPrefix(u.Path, "/dessant/buster-client/releases/download/v"+version+"/") {
			return "", "", errors.New("release asset URL does not match requested version")
		}
		return asset.BrowserDownloadURL, strings.ToLower(digest), nil
	}
	return "", "", errors.New("requested release asset was not found")
}

// NewGitHubMetadataClient creates a bounded client that only follows HTTPS
// redirects to api.github.com.
func NewGitHubMetadataClient() *http.Client {
	return &http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 || req.URL.Scheme != "https" || req.URL.Hostname() != "api.github.com" || req.URL.Port() != "" || req.URL.User != nil {
				return errors.New("refusing untrusted GitHub metadata redirect")
			}
			return nil
		},
	}
}
