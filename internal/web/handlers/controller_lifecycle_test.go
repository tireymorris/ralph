package handlers

import (
	"testing"
	"time"

	"github.com/tireymorris/ralph/internal/shared/config"
	"github.com/tireymorris/ralph/internal/shared/runner"
	runctrl "github.com/tireymorris/ralph/internal/web/runner"
	"github.com/tireymorris/ralph/internal/web/runs"
	"github.com/tireymorris/ralph/internal/workflow/events"
)

func TestReleaseControllerOnCompleted(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	reg := runs.NewRegistry()
	runID := "run-complete"
	if err := reg.Register(&runs.Run{
		ID:        runID,
		WorkDir:   workDir,
		Prompt:    "goal",
		Status:    "running",
		Phase:     "implement",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}

	api := NewAPI(cfg, reg)
	ctrl := runctrl.NewControllerWithRunner(cfg, reg, runID, &runner.NoopRunner{Runner: "mock", Command: "mock"})
	api.registerController(runID, ctrl)

	ctrl.EmitEvent(events.EventCompleted{})

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		api.mu.Lock()
		_, ok := api.controllers[runID]
		api.mu.Unlock()
		if !ok {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("controller still registered after EventCompleted")
}
