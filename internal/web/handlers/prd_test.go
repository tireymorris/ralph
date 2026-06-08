package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/web/handlers"
	"ralph/internal/web/runs"
)

func TestGetRunPRDReturnsStories(t *testing.T) {
	workDir := t.TempDir()
	prdJSON := `{
  "version": 1,
  "project_name": "test",
  "stories": [
    {"id": "s1", "title": "a", "description": "d", "acceptance_criteria": ["c"], "priority": 1, "passes": false}
  ]
}`
	if err := os.WriteFile(filepath.Join(workDir, "prd.json"), []byte(prdJSON), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	reg := runs.NewRegistry()
	now := time.Now()
	if err := reg.Register(&runs.Run{
		ID:        "run-prd",
		WorkDir:   workDir,
		Prompt:    "goal",
		Status:    "waiting_review",
		Phase:     "review",
		CreatedAt: now,
		UpdatedAt: now,
		PRDPath:   "prd.json",
	}); err != nil {
		t.Fatal(err)
	}

	api := handlers.NewAPI(cfg, reg)
	req := httptest.NewRequest(http.MethodGet, "/api/runs/run-prd/prd", nil)
	req.SetPathValue("id", "run-prd")
	rec := httptest.NewRecorder()
	api.GetRunPRD(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body struct {
		Stories []struct {
			ID                 string   `json:"id"`
			AcceptanceCriteria []string `json:"acceptance_criteria"`
			Priority           int      `json:"priority"`
		} `json:"stories"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if len(body.Stories) != 1 {
		t.Fatalf("len(stories) = %d, want 1", len(body.Stories))
	}
	if body.Stories[0].ID != "s1" {
		t.Fatalf("story id = %q, want s1", body.Stories[0].ID)
	}
	if len(body.Stories[0].AcceptanceCriteria) != 1 || body.Stories[0].AcceptanceCriteria[0] != "c" {
		t.Fatalf("acceptance_criteria = %#v, want [c]", body.Stories[0].AcceptanceCriteria)
	}
	if body.Stories[0].Priority != 1 {
		t.Fatalf("priority = %d, want 1", body.Stories[0].Priority)
	}
}
