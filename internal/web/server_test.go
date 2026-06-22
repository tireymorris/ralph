package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/runner"
	"ralph/internal/shared/testgit"
	"ralph/internal/web/runs"
)

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
	testgit.InitRepo(t, workDir)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir

	req := httptest.NewRequest(http.MethodPost, "/api/runs", strings.NewReader(`{"prompt":"goal"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h, api, err := buildHandler(cfg)
	if err != nil {
		t.Fatalf("buildHandler: %v", err)
	}
	api.SetRunnerFactory(func(*config.Config) (runner.RunnerInterface, error) {
		return runner.NoopRunner{Runner: "mock", Command: "mock"}, nil
	})
	t.Cleanup(func() {
		api.ReleaseAllControllers()
		_ = os.RemoveAll(filepath.Join(workDir, ".ralph"))
	})

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var created map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	waitForRunTerminal(t, h, created["id"])
}

func waitForRunTerminal(t *testing.T, h http.Handler, id string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/runs/"+id, nil))
		if rec.Code == http.StatusOK {
			var run struct {
				Status string `json:"status"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &run); err == nil && runs.IsTerminalStatus(run.Status) {
				return
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("run did not reach terminal status")
}
