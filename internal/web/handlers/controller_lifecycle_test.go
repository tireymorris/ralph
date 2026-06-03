package handlers

import (
	"context"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/runner"
	runctrl "ralph/internal/web/runner"
	"ralph/internal/web/runs"
	"ralph/internal/workflow/events"
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
	ctrl := runctrl.NewControllerWithRunner(cfg, reg, runID, &lifecycleNoopRunner{})
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

type lifecycleNoopRunner struct{}

func (lifecycleNoopRunner) Run(context.Context, string, chan<- runner.OutputLine) error {
	return nil
}
func (lifecycleNoopRunner) RunnerName() string        { return "mock" }
func (lifecycleNoopRunner) CommandName() string       { return "mock" }
func (lifecycleNoopRunner) IsInternalLog(string) bool { return false }
