package handlers_test

import (
	"net/http"
	"testing"

	"ralph/internal/web/runs"
)

func TestResumeRunRejectsNonGitWorkdir(t *testing.T) {
	workDir := t.TempDir()
	api, _ := setupTestAPI(t, &runs.Run{
		ID:      "run-bad-wd",
		WorkDir: workDir,
		Prompt:  "goal",
		Status:  "running",
		Phase:   "implement",
	})

	rec := postJSON(t, api.ResumeRun, "/api/runs/run-bad-wd/resume", "run-bad-wd", "{}")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}
