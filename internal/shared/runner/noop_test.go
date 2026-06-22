package runner

import (
	"context"
	"testing"
	"time"
)

func TestNoopRunnerRunCompletesWithoutBlocking(t *testing.T) {
	r := NoopRunner{}
	outputCh := make(chan OutputLine)

	done := make(chan error, 1)
	go func() {
		done <- r.Run(context.Background(), "prompt", outputCh)
	}()

	close(outputCh)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v, want nil", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Run() blocked on closed output channel")
	}
}

func TestNoopRunnerDefaultNames(t *testing.T) {
	r := NoopRunner{}

	if got := r.RunnerName(); got != "noop" {
		t.Fatalf("RunnerName() = %q, want noop", got)
	}
	if got := r.CommandName(); got != "noop" {
		t.Fatalf("CommandName() = %q, want noop", got)
	}
}

func TestNoopRunnerNameOverrides(t *testing.T) {
	r := NoopRunner{Runner: "mock", Command: "mock"}

	if got := r.RunnerName(); got != "mock" {
		t.Fatalf("RunnerName() = %q, want mock", got)
	}
	if got := r.CommandName(); got != "mock" {
		t.Fatalf("CommandName() = %q, want mock", got)
	}
}
