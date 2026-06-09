package runner

import (
	"context"
	"strings"
	"testing"

	"ralph/internal/shared/config"
)

func TestNewCursorAgent(t *testing.T) {
	cfg := &config.Config{Runner: "cursor"}
	r := NewCursorAgent(cfg)

	if r == nil {
		t.Fatal("NewCursorAgent() returned nil")
	}
	if r.CmdFunc == nil {
		t.Error("CmdFunc should not be nil")
	}
}

func TestCursorAgentIsInternalLog(t *testing.T) {
	cfg := &config.Config{Runner: "cursor"}
	r := NewCursorAgent(cfg)

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

func TestParseCursorStreamJSON_AssistantText(t *testing.T) {
	line := `{"type":"assistant","message":{"content":[{"type":"text","text":"hello world"}]}}`
	lines := parseCursorStreamJSON(line)
	if len(lines) != 1 {
		t.Fatalf("expected 1 output, got %d", len(lines))
	}
	if lines[0].Text != "hello world" {
		t.Errorf("expected text 'hello world', got %q", lines[0].Text)
	}
	if lines[0].IsErr {
		t.Error("assistant text should not be an error")
	}
}

func TestParseCursorStreamJSON_ToolUse(t *testing.T) {
	line := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"bash"}]}}`
	lines := parseCursorStreamJSON(line)
	if len(lines) != 1 {
		t.Fatalf("expected 1 output, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0].Text, "Using tool:") {
		t.Errorf("expected text starting with 'Using tool:', got %q", lines[0].Text)
	}
}

func TestParseCursorStreamJSON_ResultSuccess(t *testing.T) {
	line := `{"type":"result","subtype":"success"}`
	lines := parseCursorStreamJSON(line)
	if len(lines) != 1 {
		t.Fatalf("expected 1 output, got %d", len(lines))
	}
	if !lines[0].Verbose {
		t.Error("result success should have Verbose=true")
	}
	if lines[0].IsErr {
		t.Error("result success should not be an error")
	}
}

func TestParseCursorStreamJSON_ResultError(t *testing.T) {
	line := `{"type":"result","subtype":"error"}`
	lines := parseCursorStreamJSON(line)
	if len(lines) != 1 {
		t.Fatalf("expected 1 output, got %d", len(lines))
	}
	if !lines[0].IsErr {
		t.Error("result error should have IsErr=true")
	}
}

func TestParseCursorStreamJSON_UnknownType(t *testing.T) {
	lines := parseCursorStreamJSON(`{"type":"unknown_event"}`)
	if lines != nil {
		t.Errorf("unknown event type should return nil, got %v", lines)
	}
}

func TestParseCursorStreamJSON_MalformedJSON(t *testing.T) {
	lines := parseCursorStreamJSON("not json at all")
	if len(lines) != 1 {
		t.Fatalf("expected 1 output, got %d", len(lines))
	}
	if !lines[0].Verbose {
		t.Error("malformed JSON line should have Verbose=true")
	}
}

func TestCursorAgentRunArgs(t *testing.T) {
	cfg := &config.Config{Runner: "cursor"}
	r := NewCursorAgent(cfg)

	var capturedArgs []string
	mock := &mockCmd{stdout: "", stderr: ""}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		capturedArgs = args
		return mock
	}

	if err := r.Run(context.Background(), "do something", nil); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	expectedArgs := []string{"--print", "--output-format", "stream-json", "--trust", "--yolo"}
	assertArgsEqual(t, capturedArgs, expectedArgs)
	assertNoModelSelectionArgs(t, capturedArgs)
	assertPromptDeliveredViaStdin(t, mock, "do something")
}

func TestCursorAgentSupportsLargePrompts(t *testing.T) {
	cfg := &config.Config{Runner: "cursor"}
	r := NewCursorAgent(cfg)

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
