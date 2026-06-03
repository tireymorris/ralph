package workflow

import (
	"context"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
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
