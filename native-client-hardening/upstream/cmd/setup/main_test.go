package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func guardedRequest(method, host, remote, origin, contentType, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "http://"+host+"/api/v1/setup/install", strings.NewReader(body))
	req.Host = host
	req.RemoteAddr = remote
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	res := httptest.NewRecorder()
	requestGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(res, req)
	return res
}

func TestRequestGuardRejectsUntrustedRequests(t *testing.T) {
	previous := expectedHost
	expectedHost = "127.0.0.1:45678"
	defer func() { expectedHost = previous }()

	chromeOrigin := "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	tests := []struct {
		name, method, host, remote, origin, contentType, body string
		want int
	}{
		{"trusted extension form POST", "POST", expectedHost, "127.0.0.1:1234", chromeOrigin, "application/x-www-form-urlencoded", "session=token", http.StatusNoContent},
		{"top-level GET without Origin", "GET", expectedHost, "127.0.0.1:1234", "", "", "", http.StatusNoContent},
		{"wrong Host", "GET", "localhost:45678", "127.0.0.1:1234", "", "", "", http.StatusForbidden},
		{"non-loopback peer", "GET", expectedHost, "192.0.2.10:1234", "", "", "", http.StatusForbidden},
		{"hostile Origin", "POST", expectedHost, "127.0.0.1:1234", "https://attacker.example", "application/x-www-form-urlencoded", "session=token", http.StatusForbidden},
		{"missing mutation Origin", "POST", expectedHost, "127.0.0.1:1234", "", "application/x-www-form-urlencoded", "session=token", http.StatusForbidden},
		{"JSON content type", "POST", expectedHost, "127.0.0.1:1234", chromeOrigin, "application/json", "{}", http.StatusBadRequest},
		{"oversized request", "POST", expectedHost, "127.0.0.1:1234", chromeOrigin, "application/x-www-form-urlencoded", strings.Repeat("x", 16*1024+1), http.StatusBadRequest},
		{"unsupported method", "PUT", expectedHost, "127.0.0.1:1234", chromeOrigin, "application/x-www-form-urlencoded", "session=token", http.StatusForbidden},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := guardedRequest(tc.method, tc.host, tc.remote, tc.origin, tc.contentType, tc.body)
			if res.Code != tc.want {
				t.Fatalf("status = %d, want %d; body=%q", res.Code, tc.want, res.Body.String())
			}
			if res.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("setup responses must not be cached")
			}
		})
	}
}
