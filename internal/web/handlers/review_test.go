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

func TestReviewNotWaitingReturns409(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	reg := runs.NewRegistry()

	runID := "run-not-review"
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
	req := httptest.NewRequest(http.MethodPost, "/api/runs/"+runID+"/review",
		strings.NewReader(`{"action":"approve"}`))
	req.SetPathValue("id", runID)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	api.ReviewRun(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestReviewReviseEmptyCritiqueReturns400(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	reg := runs.NewRegistry()

	runID := "run-review"
	if err := reg.Register(&runs.Run{
		ID:        runID,
		WorkDir:   workDir,
		Prompt:    "goal",
		Status:    "waiting_review",
		Phase:     "review",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		PRDPath:   "prd.json",
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	api := handlers.NewAPI(cfg, reg)
	req := httptest.NewRequest(http.MethodPost, "/api/runs/"+runID+"/review",
		strings.NewReader(`{"action":"revise","critique":""}`))
	req.SetPathValue("id", runID)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	api.ReviewRun(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}
