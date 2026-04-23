package runner

import (
	"context"
	"strings"
	"testing"

	"ralph/internal/config"
)

func TestNewCursorAgent(t *testing.T) {
	cfg := &config.Config{Model: "cursor-agent/sonnet-4"}
	r := NewCursorAgent(cfg)

	if r == nil {
		t.Fatal("NewCursorAgent() returned nil")
	}
	if r.CmdFunc == nil {
		t.Error("CmdFunc should not be nil")
	}
}

func TestCursorAgentIsInternalLog(t *testing.T) {
	cfg := &config.Config{Model: "cursor-agent/sonnet-4"}
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
	tests := []struct {
		name         string
		model        string
		prompt       string
		expectedArgs []string
	}{
		{
			name:   "with model suffix",
			model:  "cursor-agent/sonnet-4",
			prompt: "do something",
			expectedArgs: []string{
				"--print", "--output-format", "stream-json", "--trust", "--yolo",
				"--model", "sonnet-4", "do something",
			},
		},
		{
			name:   "empty model suffix",
			model:  "cursor-agent/",
			prompt: "do something",
			expectedArgs: []string{
				"--print", "--output-format", "stream-json", "--trust", "--yolo",
				"do something",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Model: tt.model}
			r := NewCursorAgent(cfg)

			var capturedArgs []string
			mock := &mockCmd{stdout: "", stderr: ""}
			r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
				capturedArgs = args
				return mock
			}

			_ = r.Run(context.Background(), tt.prompt, nil)

			if len(capturedArgs) != len(tt.expectedArgs) {
				t.Fatalf("got args %v (len %d), want %v (len %d)",
					capturedArgs, len(capturedArgs), tt.expectedArgs, len(tt.expectedArgs))
			}
			for i, want := range tt.expectedArgs {
				if capturedArgs[i] != want {
					t.Errorf("arg[%d]: got %q, want %q", i, capturedArgs[i], want)
				}
			}
		})
	}
}
