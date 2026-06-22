package handlers_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
	runctrl "ralph/internal/web/runner"
	"ralph/internal/web/runs"
)

func TestFollowUpOnRunningReturns409(t *testing.T) {
	api, _ := setupTestAPI(t, &runs.Run{
		ID:      "run-active",
		Prompt:  "goal",
		Status:  "running",
		Phase:   "clarify",
		PRDPath: "prd.json",
	})

	rec := postJSON(t, api.FollowUpRun, "/api/runs/run-active/followup", "run-active", `{"message":"add tests"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestFollowUpCompletedUnmarksStoriesAndIncrementsVersion(t *testing.T) {
	runID := "run-done"
	api, reg := setupTestAPI(t, &runs.Run{
		ID:      runID,
		Prompt:  "goal",
		Status:  "completed",
		Phase:   "complete",
		PRDPath: "prd.json",
	})

	cfg := api.Cfg()
	cfg.PRDFile = "prd.json"

	run, _ := reg.Get(runID)
	prdData := `{
  "version": 3,
  "project_name": "Test",
  "branch_name": "feature/x",
  "stories": [
    {"id": "story-1", "title": "One", "description": "d", "slices": [{"id": "slice-1", "behavior": "a", "red_hint": "add failing test", "passes": true}], "priority": 1, "passes": true},
    {"id": "story-2", "title": "Two", "description": "d", "slices": [{"id": "slice-1", "behavior": "a", "red_hint": "add failing test", "passes": true}], "priority": 2, "passes": true}
  ]
}`
	if err := os.WriteFile(filepath.Join(run.WorkDir, "prd.json"), []byte(prdData), 0644); err != nil {
		t.Fatal(err)
	}

	implMock := &blockingImplRunner{}
	ctrl := runctrl.NewControllerWithRunner(cfg, reg, runID, implMock)
	t.Cleanup(ctrl.Cancel)
	api.SetRunnerFactory(func(*config.Config) (runner.RunnerInterface, error) { return &noopRunner{}, nil })
	api.SetController(runID, ctrl)

	rec := postJSON(t, api.FollowUpRun, "/api/runs/"+runID+"/followup", runID, `{"message":"add more tests"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		p, err := prd.Load(cfg)
		if err != nil {
			t.Fatalf("prd.Load: %v", err)
		}
		if p.Version >= 4 {
			for _, s := range p.Stories {
				if s.Passes {
					t.Fatalf("story %s still passes after follow-up", s.ID)
				}
			}
			ctrl.Cancel()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	p, _ := prd.Load(cfg)
	t.Fatalf("version = %d, want >= 4", p.Version)
}

func TestFollowUpRecreatesControllerWhenReleased(t *testing.T) {
	runID := "run-released"
	api, reg := setupTestAPI(t, &runs.Run{
		ID:      runID,
		Prompt:  "goal",
		Status:  "completed",
		Phase:   "complete",
		PRDPath: "prd.json",
	})
	api.SetRunnerFactory(func(*config.Config) (runner.RunnerInterface, error) { return &noopRunner{}, nil })

	rec := postJSON(t, api.FollowUpRun, "/api/runs/"+runID+"/followup", runID, `{"message":"tweak scope"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		got, ok := reg.Get(runID)
		if ok && got.Status != "completed" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}
