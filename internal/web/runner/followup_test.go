package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
	"ralph/internal/web/runs"
)

func TestFollowUpResumeTransitionsToImplementPhase(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:        "run-followup",
		WorkDir:   workDir,
		Prompt:    "build feature",
		Status:    "completed",
		Phase:     "complete",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		PRDPath:   "prd.json",
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	prdPath := filepath.Join(workDir, "prd.json")
	data := `{"version":1,"project_name":"Test","branch_name":"feature/x","stories":[{"id":"s1","title":"Story","description":"Do it","acceptance_criteria":["AC"],"priority":1}]}`
	if err := os.WriteFile(prdPath, []byte(data), 0644); err != nil {
		t.Fatalf("WriteFile prd: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	mock := &testRunner{
		runFunc: func(ctx context.Context, prompt string, _ chan<- runner.OutputLine) error {
			if strings.Contains(prompt, "implementation agent") {
				<-ctx.Done()
				return ctx.Err()
			}
			return nil
		},
	}
	ctrl := NewControllerWithRunner(cfg, reg, run.ID, mock)
	t.Cleanup(ctrl.Cancel)

	p, err := prd.Load(cfg)
	if err != nil {
		t.Fatalf("prd.Load: %v", err)
	}
	p.UnmarkAllStories()
	if err := prd.Save(cfg, p); err != nil {
		t.Fatalf("prd.Save: %v", err)
	}
	p, err = prd.Load(cfg)
	if err != nil {
		t.Fatalf("prd.Load after save: %v", err)
	}

	_ = reg.UpdateStatus(run.ID, "running", "followup")
	ctx := context.Background()
	ctrl.StartImplementation(ctx, p)

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		got, ok := reg.Get(run.ID)
		if ok && strings.Contains(got.Phase, "implement") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	got, _ := reg.Get(run.ID)
	t.Fatalf("registry phase = %q status = %q, want phase containing implement within 5s", got.Phase, got.Status)
}
