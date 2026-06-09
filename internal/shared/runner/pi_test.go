package runner

import (
	"context"
	"strings"
	"testing"

	"ralph/internal/shared/config"
)

func TestPiRunnerRunArgs(t *testing.T) {
	cfg := &config.Config{Runner: "pi"}
	r := NewPi(cfg)

	var capturedArgs []string
	mock := &mockCmd{stdout: "", stderr: ""}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		capturedArgs = args
		return mock
	}

	if err := r.Run(context.Background(), "test prompt", nil); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	expectedArgs := []string{"--print", "--mode", "json", "--no-session"}
	assertArgsEqual(t, capturedArgs, expectedArgs)
	assertNoModelSelectionArgs(t, capturedArgs)
	assertPromptDeliveredViaStdin(t, mock, "test prompt")
}

func TestPiRunnerSupportsLargePrompts(t *testing.T) {
	cfg := &config.Config{Runner: "pi"}
	r := NewPi(cfg)

	prompt := strings.Repeat("implement feature ", 40000)
	mock := &mockCmd{stdout: "", stderr: ""}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		for _, arg := range args {
			if strings.Contains(arg, "implement feature") {
				t.Fatal("prompt must not be passed as a CLI argument")
			}
		}
		return mock
	}

	if err := r.Run(context.Background(), prompt, nil); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	assertPromptDeliveredViaStdin(t, mock, prompt)
}
