package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
	"ralph/internal/web/runs"
	"ralph/internal/workflow/events"
)

func TestEventPRDReviewSetsCheckpoint(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:        "run-ckpt-review",
		WorkDir:   workDir,
		Prompt:    "build feature",
		Status:    "running",
		Phase:     "generate",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		PRDPath:   "prd.json",
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	prdPath := filepath.Join(workDir, "prd.json")
	data := `{"project_name":"Test","branch_name":"feature/x","stories":[{"id":"s1","title":"Story","description":"Do it","acceptance_criteria":["AC"],"priority":1}]}`
	if err := os.WriteFile(prdPath, []byte(data), 0644); err != nil {
		t.Fatalf("WriteFile prd: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	p, err := prd.Load(cfg)
	if err != nil {
		t.Fatalf("prd.Load: %v", err)
	}

	ctrl := NewControllerWithRunner(cfg, reg, run.ID, &testRunner{})
	t.Cleanup(ctrl.Cancel)

	ctrl.EmitEvent(events.EventPRDReview{PRD: p})

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		got, ok := reg.Get(run.ID)
		if ok && got.Checkpoint == runs.CheckpointPRDReview {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	got, _ := reg.Get(run.ID)
	t.Fatalf("Checkpoint = %q, want %q", got.Checkpoint, runs.CheckpointPRDReview)
}

func TestEventPRDGeneratedUpdatesRegistryPhase(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:        "run-prd-gen",
		WorkDir:   workDir,
		Prompt:    "build feature",
		Status:    "running",
		Phase:     "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		PRDPath:   "prd.json",
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	mock := &testRunner{
		runFunc: func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
			prdPath := filepath.Join(workDir, "prd.json")
			data := `{"project_name":"Generated","stories":[{"id":"1","title":"Test","description":"Desc","acceptance_criteria":["AC"],"priority":1}]}`
			return os.WriteFile(prdPath, []byte(data), 0644)
		},
	}

	ctrl := NewControllerWithRunner(cfg, reg, run.ID, mock)
	t.Cleanup(ctrl.Cancel)
	ctrl.StartNew(context.Background(), run.Prompt)

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		got, ok := reg.Get(run.ID)
		if ok && got.Phase != "" {
			ctrl.Cancel()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	got, _ := reg.Get(run.ID)
	t.Fatalf("registry phase still empty after 1s, status=%q phase=%q", got.Status, got.Phase)
}
