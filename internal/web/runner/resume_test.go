package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/runner"
	"ralph/internal/web/runs"
	"ralph/internal/workflow/events"
)

func TestForceResumeContinuesImplementation(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:        "run-resume",
		WorkDir:   workDir,
		Prompt:    "build feature",
		Status:    "implementing",
		Phase:     "implement",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		PRDPath:   "prd.json",
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	prdPath := filepath.Join(workDir, "prd.json")
	data := `{"version":1,"project_name":"Test","branch_name":"feature/x","stories":[{"id":"s1","title":"Story","description":"Do it","acceptance_criteria":["AC"],"priority":1,"passes":false}]}`
	if err := os.WriteFile(prdPath, []byte(data), 0644); err != nil {
		t.Fatalf("WriteFile prd: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	mock := &testRunner{
		runFunc: func(ctx context.Context, prompt string, _ chan<- runner.OutputLine) error {
			if strings.Contains(prompt, "You are Ralph's implementation agent") {
				<-ctx.Done()
				return ctx.Err()
			}
			return nil
		},
	}
	ctrl := NewControllerWithRunner(cfg, reg, run.ID, mock)
	t.Cleanup(func() {
		ctrl.Cancel()
		time.Sleep(100 * time.Millisecond)
	})

	ctrl.ForceResume(context.Background())

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		got, ok := reg.Get(run.ID)
		if ok && strings.Contains(got.Phase, "implement") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("run did not return to implement phase after ForceResume")
}

func TestForceResumeReloadsPRDForWaitingReview(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:        "run-review",
		WorkDir:   workDir,
		Prompt:    "build feature",
		Status:    "waiting_review",
		Phase:     "review",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		PRDPath:   "prd.json",
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	prdPath := filepath.Join(workDir, "prd.json")
	data := `{"version":1,"project_name":"Test","branch_name":"feature/x","stories":[{"id":"s1","title":"Story","description":"Do it","acceptance_criteria":["AC"],"priority":1,"passes":false}]}`
	if err := os.WriteFile(prdPath, []byte(data), 0644); err != nil {
		t.Fatalf("WriteFile prd: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	ctrl := NewControllerWithRunner(cfg, reg, run.ID, &testRunner{})
	t.Cleanup(func() {
		ctrl.Cancel()
		time.Sleep(100 * time.Millisecond)
	})

	ctrl.ForceResume(context.Background())

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		got, ok := reg.Get(run.ID)
		if ok && got.Status == "waiting_review" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("run did not return to waiting_review after ForceResume")
}

func TestForceResumeRestartsFromPromptWithoutPRD(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:        "run-restart",
		WorkDir:   workDir,
		Prompt:    "build feature",
		Status:    "running",
		Phase:     "clarify",
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

	ctrl := NewControllerWithRunner(cfg, reg, run.ID, &testRunner{})
	t.Cleanup(func() {
		ctrl.Cancel()
		time.Sleep(100 * time.Millisecond)
	})

	ctrl.ForceResume(context.Background())

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		got, ok := reg.Get(run.ID)
		if ok && got.Phase == "clarify" && got.Status == "running" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("run did not restart clarify phase after ForceResume without PRD")
}

func TestForceResumeEmitsErrorWithoutPRDOrPrompt(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:        "run-empty",
		WorkDir:   workDir,
		Prompt:    "",
		Status:    "running",
		Phase:     "clarify",
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

	ctrl := NewControllerWithRunner(cfg, reg, run.ID, &testRunner{})
	t.Cleanup(ctrl.Cancel)

	ch, unsub := ctrl.Subscribe()
	defer unsub()

	ctrl.ForceResume(context.Background())

	select {
	case ev := <-ch:
		errEv, ok := ev.(events.EventError)
		if !ok {
			t.Fatalf("event = %T, want EventError", ev)
		}
		if errEv.Err == nil || !strings.Contains(errEv.Err.Error(), "cannot resume") {
			t.Fatalf("error = %v, want cannot resume", errEv.Err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for EventError")
	}
}
