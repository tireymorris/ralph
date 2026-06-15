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

func TestCopilotRunnerIsInternalLog(t *testing.T) {
	cfg := &config.Config{Runner: "copilot"}
	r := NewCopilot(cfg)

	tests := []struct {
		line string
		want bool
	}{
		{"debug info", true},
		{"loading config", true},
		{"error: something failed", false},
		{"failed: could not connect", false},
	}

	for _, tt := range tests {
		got := r.IsInternalLog(tt.line)
		if got != tt.want {
			t.Errorf("IsInternalLog(%q) = %v, want %v", tt.line, got, tt.want)
		}
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

	expectedArgs := []string{
		"--allow-all-tools",
		"--allow-all-paths",
		"--no-ask-user",
		"--output-format", "json",
		"--autopilot",
		"--max-autopilot-continues", "50",
	}
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
	if !lines[0].Append {
		t.Error("Append should be true for message deltas")
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

func TestParseCopilotJSONL_SessionMCPServersLoaded(t *testing.T) {
	line := `{"type":"session.mcp_servers_loaded","data":{}}`
	lines := parseCopilotJSONL(line)
	if len(lines) != 1 {
		t.Fatalf("expected 1 output, got %d", len(lines))
	}
	if !lines[0].Verbose {
		t.Error("Verbose should be true")
	}
	if lines[0].IsErr {
		t.Error("IsErr should be false")
	}
}

func TestParseCopilotJSONL_AssistantTurnStart(t *testing.T) {
	line := `{"type":"assistant.turn_start","data":{}}`
	lines := parseCopilotJSONL(line)
	if len(lines) != 1 {
		t.Fatalf("expected 1 output, got %d", len(lines))
	}
	if !lines[0].Verbose {
		t.Error("Verbose should be true")
	}
}

func TestParseCopilotJSONL_ToolExecutionComplete(t *testing.T) {
	line := `{"type":"tool.execution_complete","data":{"toolName":"bash"}}`
	lines := parseCopilotJSONL(line)
	if len(lines) != 1 {
		t.Fatalf("expected 1 output, got %d", len(lines))
	}
	if !lines[0].Verbose {
		t.Error("Verbose should be true")
	}
}

func TestParseCopilotJSONL_Result(t *testing.T) {
	line := `{"type":"result","exitCode":0,"sessionId":"abc"}`
	lines := parseCopilotJSONL(line)
	if len(lines) != 1 {
		t.Fatalf("expected 1 output, got %d", len(lines))
	}
	if !lines[0].Verbose {
		t.Error("Verbose should be true")
	}
	if lines[0].Text != "exit code: 0" {
		t.Errorf("Text = %q, want %q", lines[0].Text, "exit code: 0")
	}
}

func TestParseCopilotJSONL_ResultFailure(t *testing.T) {
	line := `{"type":"result","exitCode":1,"sessionId":"abc"}`
	lines := parseCopilotJSONL(line)
	if len(lines) != 1 {
		t.Fatalf("expected 1 output, got %d", len(lines))
	}
	if lines[0].Text != "exit code: 1" {
		t.Errorf("Text = %q, want %q", lines[0].Text, "exit code: 1")
	}
}

func TestParseCopilotJSONL_ResultNestedExitCodeFallback(t *testing.T) {
	line := `{"type":"result","data":{"exitCode":2}}`
	lines := parseCopilotJSONL(line)
	if len(lines) != 1 {
		t.Fatalf("expected 1 output, got %d", len(lines))
	}
	if lines[0].Text != "exit code: 2" {
		t.Errorf("Text = %q, want %q", lines[0].Text, "exit code: 2")
	}
}

func TestParseCopilotJSONL_AssistantMessageVerbose(t *testing.T) {
	line := `{"type":"assistant.message","data":{"content":"hello from copilot"}}`
	lines := parseCopilotJSONL(line)
	if len(lines) != 1 {
		t.Fatalf("expected 1 output, got %d", len(lines))
	}
	if !lines[0].Verbose {
		t.Error("assistant.message should be verbose to avoid duplicating deltas")
	}
	if lines[0].Text != "hello from copilot" {
		t.Errorf("Text = %q, want %q", lines[0].Text, "hello from copilot")
	}
}

func TestCopilotResultExitCode(t *testing.T) {
	tests := []struct {
		top, data, want int
	}{
		{0, 0, 0},
		{1, 0, 1},
		{0, 2, 2},
	}
	for _, tt := range tests {
		if got := copilotResultExitCode(tt.top, tt.data); got != tt.want {
			t.Errorf("copilotResultExitCode(%d, %d) = %d, want %d", tt.top, tt.data, got, tt.want)
		}
	}
}

func TestParseCopilotJSONL_UnknownEvent(t *testing.T) {
	lines := parseCopilotJSONL(`{"type":"unknown_event","data":{}}`)
	if lines != nil {
		t.Errorf("unknown event type should return nil, got %v", lines)
	}
}
