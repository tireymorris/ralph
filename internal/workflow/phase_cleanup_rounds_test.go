package workflow

import (
	"context"
	"strings"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
	"ralph/internal/shared/testgit"
)

func TestRunCleanupMultipleRoundsUntilNoProgress(t *testing.T) {
	workDir, changedFile := testgit.RepoWithWorkingTreeDiff(t)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()
	calls := 0
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		calls++
		if !strings.Contains(prompt, changedFile) {
			t.Errorf("cleanup prompt should contain changed file %q", changedFile)
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	p := &prd.PRD{Context: "ctx"}

	if err := exec.RunCleanup(context.Background(), p); err != nil {
		t.Fatalf("RunCleanup() error = %v", err)
	}

	if calls != 1 {
		t.Fatalf("runner call count = %d, want 1 when cleanup makes no file changes", calls)
	}

	evts := drainEvents(ch)
	counts := countCleanupEvents(evts)
	if counts.started != 1 || counts.completed != 1 {
		t.Errorf("expected 1 cleanup started and 1 completed, got started=%d completed=%d", counts.started, counts.completed)
	}
}

func TestRunCleanupFailsWhenTestsFail(t *testing.T) {
	workDir, _ := testgit.RepoWithWorkingTreeDiff(t)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"
	cfg.TestCommand = "exit 1"

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	p := &prd.PRD{Context: "ctx"}

	err := exec.RunCleanup(context.Background(), p)
	if err == nil {
		t.Fatal("RunCleanup() error = nil, want test gate failure")
	}

	evts := drainEvents(ch)
	counts := countCleanupEvents(evts)
	if counts.started != 1 {
		t.Errorf("EventCleanupStarted count = %d, want 1", counts.started)
	}
	if counts.completed != 0 {
		t.Errorf("EventCleanupCompleted count = %d, want 0 on test failure", counts.completed)
	}
}
