package runner

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/runner"
	"ralph/internal/web/runs"
)

func TestCancelStopsWorkflowWithinTwoSeconds(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:        "run-cancel",
		WorkDir:   workDir,
		Prompt:    "block",
		Status:    "running",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir

	var sawCtxErr atomic.Bool
	mock := &testRunner{
		runFunc: func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
			<-ctx.Done()
			sawCtxErr.Store(true)
			return ctx.Err()
		},
	}

	ctrl := NewControllerWithRunner(cfg, reg, run.ID, mock)
	ctrl.StartNew(context.Background(), run.Prompt)

	time.Sleep(20 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		ctrl.Cancel()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Cancel() did not return within 2s")
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if sawCtxErr.Load() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("workflow goroutine did not observe ctx.Err() after Cancel()")
}
