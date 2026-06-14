package runner

import (
	"context"
	"strings"
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

func TestCopilotRunnerSupportsLargePrompts(t *testing.T) {
	cfg := &config.Config{Runner: "copilot"}
	r := NewCopilot(cfg)

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

func TestParseCopilotJSONL_MessageDelta(t *testing.T) {
	line := `{"type":"assistant.message_delta","data":{"deltaContent":"hello"}}`
	lines := parseCopilotJSONL(line)
	if len(lines) != 1 {
		t.Fatalf("expected 1 output, got %d", len(lines))
	}
	if lines[0].Text != "hello" {
		t.Errorf("Text = %q, want %q", lines[0].Text, "hello")
	}
	if lines[0].IsErr {
		t.Error("IsErr should be false")
	}
	if lines[0].Verbose {
		t.Error("Verbose should be false")
	}
}

func TestParseCopilotJSONL_ToolExecutionStart(t *testing.T) {
	line := `{"type":"tool.execution_start","data":{"toolName":"bash"}}`
	lines := parseCopilotJSONL(line)
	if len(lines) != 1 {
		t.Fatalf("expected 1 output, got %d", len(lines))
	}
	if lines[0].Text != "Using tool: bash" {
		t.Errorf("Text = %q, want %q", lines[0].Text, "Using tool: bash")
	}
	if lines[0].IsErr {
		t.Error("IsErr should be false")
	}
	if lines[0].Verbose {
		t.Error("Verbose should be false")
	}
}

func TestParseCopilotJSONL_SessionError(t *testing.T) {
	line := `{"type":"session.error","data":{"message":"auth failed"}}`
	lines := parseCopilotJSONL(line)
	if len(lines) != 1 {
		t.Fatalf("expected 1 output, got %d", len(lines))
	}
	if lines[0].Text != "auth failed" {
		t.Errorf("Text = %q, want %q", lines[0].Text, "auth failed")
	}
	if !lines[0].IsErr {
		t.Error("IsErr should be true")
	}
	if lines[0].Verbose {
		t.Error("Verbose should be false")
	}
}

func TestParseCopilotJSONL_ModelCallFailure(t *testing.T) {
	line := `{"type":"model.call_failure","data":{"errorMessage":"rate limited"}}`
	lines := parseCopilotJSONL(line)
	if len(lines) != 1 {
		t.Fatalf("expected 1 output, got %d", len(lines))
	}
	if lines[0].Text != "rate limited" {
		t.Errorf("Text = %q, want %q", lines[0].Text, "rate limited")
	}
	if !lines[0].IsErr {
		t.Error("IsErr should be true")
	}
	if lines[0].Verbose {
		t.Error("Verbose should be false")
	}
}

func TestParseCopilotJSONL_MalformedJSON(t *testing.T) {
	lines := parseCopilotJSONL("not json at all")
	if len(lines) != 1 {
		t.Fatalf("expected 1 output, got %d", len(lines))
	}
	if !lines[0].Verbose {
		t.Error("malformed JSON line should have Verbose=true")
	}
}
