package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"ralph/internal/shared/config"
)

func initGitRepoInDir(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
}

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h, err := NewHandler(config.DefaultConfig())
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "ok") {
		t.Errorf("body = %q, want JSON containing ok", body)
	}
	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if payload["status"] != "ok" {
		t.Errorf("status field = %q, want ok", payload["status"])
	}
}

func TestCreateRunRouteRegistered(t *testing.T) {
	workDir := t.TempDir()
	initGitRepoInDir(t, workDir)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir

	req := httptest.NewRequest(http.MethodPost, "/api/runs", strings.NewReader(`{"prompt":"goal"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h, err := NewHandler(cfg)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}
