package handlers_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
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

	if _, err := os.Stat(cfg.PRDPath()); !os.IsNotExist(err) {
		t.Fatalf("prd.json still exists: %v", err)
	}
	ralphRuns := filepath.Join(cfg.WorkDir, ".ralph", "runs")
	if _, err := os.Stat(ralphRuns); !os.IsNotExist(err) {
		t.Fatalf(".ralph/runs still exists: %v", err)
	}
	if n := len(reg.List()); n != 0 {
		t.Fatalf("registry after clean: %d runs, want 0", n)
	}
}
