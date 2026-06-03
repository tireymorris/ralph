package workflow

import (
	"context"
	"strings"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
)

func TestRunCleanupContextCancelled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()
	exec := NewExecutorWithRunner(cfg, ch, mock)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := &prd.PRD{Context: "test context"}
	err := exec.RunCleanup(ctx, p)

	if err == nil {
		t.Fatal("RunCleanup() should return error when context is cancelled")
	}
	if err != context.Canceled {
		t.Fatalf("RunCleanup() error = %v, want context.Canceled", err)
	}

	for len(ch) > 0 {
		e := <-ch
		if _, ok := e.(EventCleanupStarted); ok {
			t.Error("EventCleanupStarted should not be emitted when context is cancelled")
		}
	}

	if mock.CallCount() != 0 {
		t.Errorf("runner should not be called, got %d calls", mock.CallCount())
	}
}

func TestRunCleanupSuccess(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		if !strings.Contains(prompt, "SOLID") {
			t.Error("cleanup prompt should contain SOLID")
		}
		if !strings.Contains(prompt, "my project context") {
			t.Error("cleanup prompt should contain the PRD context")
		}
		outputCh <- runner.OutputLine{Text: "refactoring..."}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	p := &prd.PRD{Context: "my project context"}

	err := exec.RunCleanup(context.Background(), p)
	if err != nil {
		t.Fatalf("RunCleanup() error = %v", err)
	}

	if mock.CallCount() != 1 {
		t.Fatalf("runner call count = %d, want 1", mock.CallCount())
	}

	var evts []Event
	for len(ch) > 0 {
		evts = append(evts, <-ch)
	}

	if len(evts) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(evts))
	}

	foundStarted := false
	foundCompleted := false
	foundOutput := false
	startedIdx := -1
	completedIdx := -1

	for i, e := range evts {
		switch e.(type) {
		case EventCleanupStarted:
			foundStarted = true
			startedIdx = i
		case EventCleanupCompleted:
			foundCompleted = true
			completedIdx = i
		case EventOutput:
			foundOutput = true
		}
	}

	if !foundStarted {
		t.Error("expected EventCleanupStarted to be emitted")
	}
	if !foundCompleted {
		t.Error("expected EventCleanupCompleted to be emitted")
	}
	if !foundOutput {
		t.Error("expected runner output to be forwarded as EventOutput")
	}
	if startedIdx >= completedIdx {
		t.Error("EventCleanupStarted should come before EventCleanupCompleted")
	}
}
