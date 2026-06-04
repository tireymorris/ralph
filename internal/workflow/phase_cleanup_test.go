package workflow

import (
	"context"
	"errors"
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

	if mock.CallCount() != 3 {
		t.Fatalf("runner call count = %d, want 3", mock.CallCount())
	}

	evts := drainEvents(ch)
	assertCleanupPassSequence(t, evts, 3)

	foundOutput := false
	for _, e := range evts {
		if _, ok := e.(EventOutput); ok {
			foundOutput = true
		}
	}
	if !foundOutput {
		t.Error("expected runner output to be forwarded as EventOutput")
	}
}

func TestRunCleanupRunnerError(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		return errors.New("something broke")
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	p := &prd.PRD{Context: "ctx"}

	err := exec.RunCleanup(context.Background(), p)
	if err == nil {
		t.Fatal("RunCleanup() should return error when runner fails")
	}

	if mock.CallCount() != 1 {
		t.Fatalf("runner call count = %d, want 1 on first-pass failure", mock.CallCount())
	}

	foundError := false
	foundCompleted := false
	var startedPass, startedTotal int
	for len(ch) > 0 {
		e := <-ch
		switch ev := e.(type) {
		case EventError:
			if strings.Contains(ev.Err.Error(), "cleanup") {
				foundError = true
			}
		case EventCleanupStarted:
			startedPass = ev.Pass
			startedTotal = ev.Total
		case EventCleanupCompleted:
			foundCompleted = true
		}
	}

	if startedPass != 1 {
		t.Errorf("EventCleanupStarted Pass=%d, want 1 on first-pass failure", startedPass)
	}
	if startedTotal != 3 {
		t.Errorf("EventCleanupStarted Total=%d, want 3", startedTotal)
	}

	if !foundError {
		t.Error("expected EventError with message containing 'cleanup'")
	}
	if foundCompleted {
		t.Error("EventCleanupCompleted should not be emitted on runner failure")
	}
}
