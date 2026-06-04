package handlers_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ralph/internal/clean"
	"ralph/internal/web/runs"
)

func TestCleanState(t *testing.T) {
	api, reg := setupTestAPI(t, &runs.Run{
		ID:     "active-run",
		Prompt: "goal",
		Status: "running",
		Phase:  "implement",
	})
	cfg := api.Cfg()
	if _, err := clean.SeedStateArtifacts(cfg); err != nil {
		t.Fatal(err)
	}
	if reg.List() == nil || len(reg.List()) != 1 {
		t.Fatalf("registry before clean: %d runs, want 1", len(reg.List()))
	}

	rec := postJSON(t, api.CleanState, "/api/clean", "", "{}")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body = %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body) != 0 {
		t.Fatalf("body = %v, want {}", body)
	}

	assertPathNotExist(t, cfg.PRDPath())
	assertRalphRunsRemoved(t, cfg.WorkDir)
	assertRegistryEmpty(t, reg)
}

func TestCleanState_RemoveStateFailure(t *testing.T) {
	api, reg := setupTestAPI(t, &runs.Run{
		ID:     "active-run",
		Prompt: "goal",
		Status: "running",
		Phase:  "implement",
	})
	cfg := api.Cfg()
	if _, err := clean.SeedStateArtifacts(cfg); err != nil {
		t.Fatal(err)
	}
	prdPath := cfg.PRDPath()
	if err := os.Remove(prdPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(prdPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(prdPath, "block"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(prdPath) })

	rec := postJSON(t, api.CleanState, "/api/clean", "", "{}")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "clean failed after stopping active runs") {
		t.Fatalf("body = %s, want clean failure message", rec.Body.String())
	}
	if len(reg.List()) != 1 {
		t.Fatalf("registry after failed clean: %d runs, want 1", len(reg.List()))
	}
}
