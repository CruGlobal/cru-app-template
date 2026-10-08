package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRoutes runs signed in through the dev bypass, so it sees the routes
// rather than the sign-in gate.
func TestRoutes(t *testing.T) {
	t.Setenv("IAP_AUDIENCE", "")
	t.Setenv("CRU_IAP_DEV_BYPASS_EMAIL", "dev@example.com")

	tests := []struct {
		path string
		want int
	}{
		{"/health", http.StatusOK},
		{"/up", http.StatusOK},
		{"/", http.StatusOK},
		{"/nope", http.StatusNotFound},
	}
	for _, tt := range tests {
		if rec := get(t, tt.path); rec.Code != tt.want {
			t.Errorf("GET %s = %d, want %d", tt.path, rec.Code, tt.want)
		}
	}
}

func get(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	routes().ServeHTTP(rec, req)
	return rec
}

func TestUpIsAlwaysOpen(t *testing.T) {
	t.Setenv("IAP_AUDIENCE", "")
	t.Setenv("CRU_IAP_DEV_BYPASS_EMAIL", "")

	if rec := get(t, "/up"); rec.Code != http.StatusOK {
		t.Errorf("GET /up = %d, want 200", rec.Code)
	}
}

func TestEverythingElseNeedsAnAssertionOrTheDevBypass(t *testing.T) {
	t.Setenv("IAP_AUDIENCE", "")
	t.Setenv("CRU_IAP_DEV_BYPASS_EMAIL", "")

	for _, path := range []string{"/", "/health"} {
		if rec := get(t, path); rec.Code != http.StatusUnauthorized {
			t.Errorf("GET %s = %d, want 401", path, rec.Code)
		}
	}
}

func TestDevBypassNamesTheUser(t *testing.T) {
	t.Setenv("IAP_AUDIENCE", "")
	t.Setenv("CRU_IAP_DEV_BYPASS_EMAIL", "dev@example.com")

	rec := get(t, "/")
	if rec.Code != http.StatusOK || rec.Body.String() != "Hello, dev@example.com 👋\n" {
		t.Errorf("GET / = %d %q, want 200 greeting dev@example.com", rec.Code, rec.Body.String())
	}
}

func TestIAPAudienceIgnoresTheDevBypass(t *testing.T) {
	t.Setenv("IAP_AUDIENCE", "/projects/1/global/backendServices/2")
	t.Setenv("CRU_IAP_DEV_BYPASS_EMAIL", "dev@example.com")

	if rec := get(t, "/"); rec.Code != http.StatusUnauthorized {
		t.Errorf("GET / = %d, want 401", rec.Code)
	}
	if rec := get(t, "/up"); rec.Code != http.StatusOK {
		t.Errorf("GET /up = %d, want 200", rec.Code)
	}
}
