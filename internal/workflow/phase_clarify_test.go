package workflow

import (
	"context"
	"strings"
	"testing"
	"time"

	"ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/runner"
)

func TestRunClarifyAutoApproveSkipsRunner(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	cfg.AutoApprove = true

	eventsCh := make(chan Event, 10)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		t.Fatal("clarify runner should not be invoked when auto-approve is enabled")
		return nil
	}
	exec := NewExecutorWithRunner(cfg, eventsCh, mock)

	done := make(chan struct {
		answers []prompt.QuestionAnswer
		err     error
	}, 1)
	go func() {
		answers, err := exec.RunClarify(context.Background(), "build something")
		done <- struct {
			answers []prompt.QuestionAnswer
			err     error
		}{answers: answers, err: err}
	}()

	select {
	case result := <-done:
		if result.err != nil {
			t.Fatalf("RunClarify() error = %v, want nil", result.err)
		}
		if result.answers != nil {
			t.Fatalf("RunClarify() answers = %v, want nil", result.answers)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("RunClarify() did not return within 100ms")
	}

	if got := mock.CallCount(); got != 0 {
		t.Fatalf("runner calls = %d, want 0", got)
	}

	select {
	case event := <-eventsCh:
		output, ok := event.(EventOutput)
		if !ok {
			t.Fatalf("event = %T, want EventOutput", event)
		}
		if !strings.Contains(output.Output.Text, "skipping clarification") {
			t.Fatalf("EventOutput text = %q, want it to contain skipping clarification", output.Output.Text)
		}
	default:
		t.Fatal("expected EventOutput for skipped clarification")
	}
}
