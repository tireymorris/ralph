package workflow

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/prd/prdtest"
	"ralph/internal/shared/runner"
	"ralph/internal/shared/runstate"
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

func TestRunCleanupRecoversWhenAutoApproveAndTestsFail(t *testing.T) {
	workDir, _ := testgit.RepoWithWorkingTreeDiff(t)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"
	cfg.AutoApprove = true
	cfg.TestCommand = finalGateTestCommand

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Context:     "ctx",
		Stories: []*prd.Story{
			{ID: "story-1", Title: "One", Description: "d", Slices: prdtest.Slices("a"), Priority: 1, Passes: true},
		},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("save PRD: %v", err)
	}

	greetDir := filepath.Join(workDir, "pkg", "greet")
	if err := os.MkdirAll(greetDir, 0o755); err != nil {
		t.Fatal(err)
	}

	ch := make(chan Event, 100)
	mock := newMockRunner()
	recoveryCalls := 0
	mock.runFunc = func(_ context.Context, p string, _ chan<- runner.OutputLine) error {
		if isRecoveryPrompt(p) {
			recoveryCalls++
			return os.WriteFile(filepath.Join(greetDir, "greet.go"), []byte(`package greet

func Hello() string { return "hello" }
`), 0o644)
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	if err := exec.RunCleanup(context.Background(), testPRD); err != nil {
		t.Fatalf("RunCleanup() error = %v", err)
	}
	if recoveryCalls != 1 {
		t.Fatalf("recovery runner calls = %d, want 1", recoveryCalls)
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
	cfg.AutoApprove = false
	cfg.TestCommand = finalGateTestCommand

	ch := make(chan Event, 100)
	mock := newMockRunner()
	recoveryCalls := 0
	mock.runFunc = func(_ context.Context, p string, _ chan<- runner.OutputLine) error {
		if isRecoveryPrompt(p) {
			recoveryCalls++
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	p := &prd.PRD{Context: "ctx"}

	err := exec.RunCleanup(context.Background(), p)
	if err == nil {
		t.Fatal("RunCleanup() error = nil, want test gate failure")
	}
	if recoveryCalls != 0 {
		t.Fatalf("recovery runner calls = %d, want 0", recoveryCalls)
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

func TestRunCleanupResetsExhaustedRecoveryBudget(t *testing.T) {
	workDir, _ := testgit.RepoWithWorkingTreeDiff(t)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"
	cfg.AutoApprove = true
	cfg.TestCommand = finalGateTestCommand

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Context:     "ctx",
		Stories: []*prd.Story{
			{ID: "story-1", Title: "One", Description: "d", Slices: prdtest.Slices("a"), Priority: 1, Passes: true},
		},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("save PRD: %v", err)
	}

	greetDir := filepath.Join(workDir, "pkg", "greet")
	if err := os.MkdirAll(greetDir, 0o755); err != nil {
		t.Fatal(err)
	}

	loop := NewFileReviewLoop(workDir, runstate.LocalRunID)
	if err := loop.Apply(ReviewLoopUpdate{RecoveryAttempts: constants.MaxRecoveryAttempts}); err != nil {
		t.Fatalf("Apply() err = %v", err)
	}

	ch := make(chan Event, 100)
	mock := newMockRunner()
	recoveryCalls := 0
	mock.runFunc = func(_ context.Context, p string, _ chan<- runner.OutputLine) error {
		if isRecoveryPrompt(p) {
			recoveryCalls++
			return os.WriteFile(filepath.Join(greetDir, "greet.go"), []byte(`package greet

func Hello() string { return "hello" }
`), 0o644)
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	exec.SetReviewLoop(runstate.LocalRunID, loop)
	if err := exec.RunCleanup(context.Background(), testPRD); err != nil {
		t.Fatalf("RunCleanup() error = %v", err)
	}
	if recoveryCalls != 1 {
		t.Fatalf("recovery runner calls = %d, want 1", recoveryCalls)
	}
}
