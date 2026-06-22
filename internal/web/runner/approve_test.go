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
	"ralph/internal/workflow/events"
)

func TestApproveReviewTransitionsToImplementing(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:        "run-approve",
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
	data := `{"project_name":"Test","branch_name":"feature/x","stories":[{"id":"s1","title":"Story","description":"Do it","slices":[{"id":"slice-1","behavior":"AC","red_hint":"add failing test","passes":false}],"priority":1,"passes":false}]}`
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
	t.Cleanup(ctrl.Cancel)

	ctrl.EmitEvent(events.EventPRDReview{PRD: p})

	syncDeadline := time.Now().Add(time.Second)
	for time.Now().Before(syncDeadline) {
		got, ok := reg.Get(run.ID)
		if ok && got.Status == "waiting_review" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if err := ctrl.ApproveReview(context.Background()); err != nil {
		t.Fatalf("ApproveReview() error = %v", err)
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		got, ok := reg.Get(run.ID)
		if ok && strings.Contains(got.Phase, "implement") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	got, _ := reg.Get(run.ID)
	t.Fatalf("registry phase = %q status = %q, want phase containing implement within 1s", got.Phase, got.Status)
}
