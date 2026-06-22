package handlers_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/runner"
	"ralph/internal/web/runs"
)

func TestResumeRunOnRunningReturnsAccepted(t *testing.T) {
	api, _ := setupTestAPI(t, &runs.Run{
		ID:      "run-active",
		Prompt:  "goal",
		Status:  "implementing",
		Phase:   "implement",
		PRDPath: "prd.json",
	})
	api.SetRunnerFactory(func(*config.Config) (runner.RunnerInterface, error) {
		return noopRunner{}, nil
	})
	t.Cleanup(func() {
		api.ReleaseAllControllers()
		time.Sleep(100 * time.Millisecond)
	})

	cfg := api.Cfg()
	prdPath := filepath.Join(cfg.WorkDir, "prd.json")
	data := `{"version":1,"project_name":"Test","branch_name":"feature/x","stories":[{"id":"s1","title":"Story","description":"Do it","slices":[{"id":"slice-1","behavior":"AC","red_hint":"add failing test","passes":false}],"priority":1,"passes":false}]}`
	if err := os.WriteFile(prdPath, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	rec := postJSON(t, api.ResumeRun, "/api/runs/run-active/resume", "run-active", "{}")
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
}

func TestResumeRunNotFoundReturns404(t *testing.T) {
	api, _ := setupTestAPI(t)

	rec := postJSON(t, api.ResumeRun, "/api/runs/missing/resume", "missing", "{}")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestResumeRunOnCompletedReturns409(t *testing.T) {
	api, _ := setupTestAPI(t, &runs.Run{
		ID:     "run-done",
		Prompt: "goal",
		Status: "completed",
		Phase:  "complete",
	})

	rec := postJSON(t, api.ResumeRun, "/api/runs/run-done/resume", "run-done", "{}")
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}
