package runner

import (
	"context"
	"testing"

	"ralph/internal/shared/config"
)

func TestNewCopilot(t *testing.T) {
	cfg := &config.Config{Runner: "copilot"}
	r := NewCopilot(cfg)

	if r == nil {
		t.Fatal("NewCopilot() returned nil")
	}
	if r.CmdFunc == nil {
		t.Error("CmdFunc should not be nil")
	}
}

func TestCopilotRunnerNames(t *testing.T) {
	cfg := &config.Config{Runner: "copilot"}
	r := NewCopilot(cfg)

	if r.RunnerName() != "copilot" {
		t.Errorf("RunnerName() = %q, want %q", r.RunnerName(), "copilot")
	}
	if r.CommandName() != "copilot" {
		t.Errorf("CommandName() = %q, want %q", r.CommandName(), "copilot")
	}
}

func TestCopilotRunnerRunArgs(t *testing.T) {
	cfg := &config.Config{Runner: "copilot"}
	r := NewCopilot(cfg)

	var capturedName string
	var capturedArgs []string
	mock := &mockCmd{stdout: "", stderr: ""}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		capturedName = name
		capturedArgs = args
		return mock
	}

	if err := r.Run(context.Background(), "do something", nil); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if capturedName != "copilot" {
		t.Errorf("command name = %q, want %q", capturedName, "copilot")
	}

	expectedArgs := []string{"--allow-all-tools", "--allow-all-paths", "--no-ask-user", "--output-format", "json", "--autopilot"}
	assertArgsEqual(t, capturedArgs, expectedArgs)
	assertNoModelSelectionArgs(t, capturedArgs)
	assertPromptDeliveredViaStdin(t, mock, "do something")
}
