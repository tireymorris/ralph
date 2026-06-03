package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/web/handlers"
	"ralph/internal/web/runs"
)

func TestClarifyNotWaitingReturns409(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	reg := runs.NewRegistry()

	runID := "run-not-waiting"
	if err := reg.Register(&runs.Run{
		ID:        runID,
		WorkDir:   workDir,
		Prompt:    "goal",
		Status:    "running",
		Phase:     "clarify",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		PRDPath:   "prd.json",
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	api := handlers.NewAPI(cfg, reg)
	req := httptest.NewRequest(http.MethodPost, "/api/runs/"+runID+"/clarify",
		strings.NewReader(`{"answers":[{"question":"Q?","answer":"A"}]}`))
	req.SetPathValue("id", runID)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	api.ClarifyRun(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}
