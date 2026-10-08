package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoutes(t *testing.T) {
	t.Parallel()

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
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			routes().ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Errorf("GET %s = %d, want %d", tt.path, rec.Code, tt.want)
			}
		})
	}
}

func get(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	routes().ServeHTTP(rec, req)
	return rec
}

func TestIAPGate(t *testing.T) {
	t.Setenv("IAP_AUDIENCE", "/projects/1/global/backendServices/2")
	t.Setenv("CRU_IAP_DEV_BYPASS_EMAIL", "dev@example.com") // must not open the gate

	if rec := get(t, "/up"); rec.Code != http.StatusOK {
		t.Errorf("GET /up = %d, want 200", rec.Code)
	}
	for _, path := range []string{"/", "/health"} {
		if rec := get(t, path); rec.Code != http.StatusUnauthorized {
			t.Errorf("GET %s without an assertion = %d, want 401", path, rec.Code)
		}
	}
}

func TestNoGateWithoutAudience(t *testing.T) {
	t.Setenv("IAP_AUDIENCE", "")
	t.Setenv("CRU_IAP_DEV_BYPASS_EMAIL", "dev@example.com")

	rec := get(t, "/")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "dev@example.com") {
		t.Errorf("GET / = %d %q, want 200 greeting dev@example.com", rec.Code, rec.Body.String())
	}
}
