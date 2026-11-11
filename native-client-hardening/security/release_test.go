package security

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func metadataClient(status int, body string) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "api.github.com" || r.URL.Path != "/repos/dessant/buster-client/releases/tags/v1.2.3" {
			return nil, errors.New("unexpected metadata request URL")
		}
		return &http.Response{
			StatusCode: status,
			Status:     http.StatusText(status),
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    r,
		}, nil
	})}
}

func TestReleaseAssetSHA256(t *testing.T) {
	digest := strings.Repeat("a", 64)
	body := `{"tag_name":"v1.2.3","assets":[{"name":"asset.zip","digest":"sha256:` + digest + `","browser_download_url":"https://github.com/dessant/buster-client/releases/download/v1.2.3/asset.zip"}]}`
	url, gotDigest, err := ReleaseAssetSHA256(metadataClient(http.StatusOK, body), "1.2.3", "asset.zip")
	if err != nil {
		t.Fatalf("ReleaseAssetSHA256 returned error: %v", err)
	}
	if url != "https://github.com/dessant/buster-client/releases/download/v1.2.3/asset.zip" {
		t.Fatalf("unexpected asset URL: %s", url)
	}
	if gotDigest != digest {
		t.Fatalf("unexpected digest: %s", gotDigest)
	}
}

func TestReleaseAssetSHA256FailsClosedWithoutDigest(t *testing.T) {
	body := `{"tag_name":"v1.2.3","assets":[{"name":"asset.zip","browser_download_url":"https://github.com/dessant/buster-client/releases/download/v1.2.3/asset.zip"}]}`
	if _, _, err := ReleaseAssetSHA256(metadataClient(http.StatusOK, body), "1.2.3", "asset.zip"); err == nil {
		t.Fatal("expected missing digest to be rejected")
	}
}

func TestReleaseAssetSHA256RejectsMismatchedTagAndBadVersion(t *testing.T) {
	body := `{"tag_name":"v9.9.9","assets":[]}`
	if _, _, err := ReleaseAssetSHA256(metadataClient(http.StatusOK, body), "1.2.3", "asset.zip"); err == nil {
		t.Fatal("expected mismatched tag to be rejected")
	}
	if _, _, err := ReleaseAssetSHA256(metadataClient(http.StatusOK, body), "../latest", "asset.zip"); err == nil {
		t.Fatal("expected invalid version to be rejected")
	}
}

func TestReleaseAssetSHA256RejectsNonSuccessStatus(t *testing.T) {
	if _, _, err := ReleaseAssetSHA256(metadataClient(http.StatusNotFound, "{}"), "1.2.3", "asset.zip"); err == nil {
		t.Fatal("expected non-200 response to be rejected")
	}
}
