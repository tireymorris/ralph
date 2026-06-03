package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ralph/internal/shared/config"
)

func mustNewHandler(t *testing.T) http.Handler {
	t.Helper()
	h, err := NewHandler(config.DefaultConfig())
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}
	return h
}

func TestServeIndexHTML(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mustNewHandler(t).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	if !strings.Contains(rec.Body.String(), "/assets/") {
		t.Errorf("body should reference /assets/, got:\n%s", rec.Body.String())
	}
}

func TestServeSPARouteFallback(t *testing.T) {
	for _, path := range []string{"/runs/abc123", "/new"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			mustNewHandler(t).ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", path, rec.Code, http.StatusOK)
			}
			if !strings.Contains(rec.Body.String(), "/assets/") {
				t.Errorf("body should reference /assets/, got:\n%s", rec.Body.String())
			}
		})
	}
}

func TestServeAssetNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/assets/missing.js", nil)
	rec := httptest.NewRecorder()

	mustNewHandler(t).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /assets/missing.js status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
