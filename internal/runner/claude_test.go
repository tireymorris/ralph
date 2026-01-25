package runner

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"ralph/internal/config"
)

func TestNewClaude(t *testing.T) {
	cfg := &config.Config{Model: "claude-code/sonnet"}
	r := NewClaude(cfg)

	if r == nil {
		t.Fatal("NewClaude() returned nil")
	}
	if r.cfg != cfg {
		t.Error("NewClaude() did not set config correctly")
	}
	if r.CmdFunc == nil {
		t.Error("CmdFunc should not be nil")
	}
}

func TestClaudeRunWithModel(t *testing.T) {
	cfg := &config.Config{Model: "claude-code/sonnet"}
	r := NewClaude(cfg)

	var capturedArgs []string
	mock := &mockCmd{stdout: "output line", stderr: ""}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		capturedArgs = args
		return mock
	}

	err := r.Run(context.Background(), "test prompt", nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	expectedArgs := []string{"--print", "--verbose", "--output-format", "stream-json", "--dangerously-skip-permissions", "--model", "sonnet", "test prompt"}
	if len(capturedArgs) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d", len(expectedArgs), len(capturedArgs))
	}
	for i, expected := range expectedArgs {
		if capturedArgs[i] != expected {
			t.Errorf("Arg %d: expected %q, got %q", i, expected, capturedArgs[i])
		}
	}
}

func TestClaudeRunNoModel(t *testing.T) {
	cfg := &config.Config{Model: ""}
	r := NewClaude(cfg)

	var capturedArgs []string
	mock := &mockCmd{}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		capturedArgs = args
		return mock
	}

	r.Run(context.Background(), "test", nil)

	expectedArgs := []string{"--print", "--verbose", "--output-format", "stream-json", "--dangerously-skip-permissions", "test"}
	if len(capturedArgs) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d", len(expectedArgs), len(capturedArgs))
	}
	for i, expected := range expectedArgs {
		if capturedArgs[i] != expected {
			t.Errorf("Arg %d: expected %q, got %q", i, expected, capturedArgs[i])
		}
	}
}

func TestClaudeRunWithOutputChannel(t *testing.T) {
	cfg := &config.Config{Model: "claude-code/haiku"}
	r := NewClaude(cfg)

	mock := &mockCmd{stdout: "line1\nline2", stderr: "err1"}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	outputCh := make(chan OutputLine, 100)
	err := r.Run(context.Background(), "test", outputCh)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	close(outputCh)
	var lines []OutputLine
	for line := range outputCh {
		lines = append(lines, line)
	}

	if len(lines) < 1 {
		t.Fatal("Expected at least one output line")
	}
}

func TestClaudeRunStdoutError(t *testing.T) {
	cfg := &config.Config{}
	r := NewClaude(cfg)

	mock := &mockCmd{stdoutErr: errors.New("stdout error")}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	err := r.Run(context.Background(), "test", nil)
	if err == nil {
		t.Error("Run() should error on stdout failure")
	}
	if !strings.Contains(err.Error(), "stdout pipe") {
		t.Errorf("Expected stdout pipe error, got %v", err)
	}
}

func TestClaudeRunStderrError(t *testing.T) {
	cfg := &config.Config{}
	r := NewClaude(cfg)

	mock := &mockCmd{stderrErr: errors.New("stderr error")}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	err := r.Run(context.Background(), "test", nil)
	if err == nil {
		t.Error("Run() should error on stderr failure")
	}
	if !strings.Contains(err.Error(), "stderr pipe") {
		t.Errorf("Expected stderr pipe error, got %v", err)
	}
}

func TestClaudeRunStartError(t *testing.T) {
	cfg := &config.Config{}
	r := NewClaude(cfg)

	mock := &mockCmd{startErr: errors.New("start error")}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	err := r.Run(context.Background(), "test", nil)
	if err == nil {
		t.Error("Run() should error on start failure")
	}
	if !strings.Contains(err.Error(), "start claude") {
		t.Errorf("Expected start error, got %v", err)
	}
}

func TestClaudeRunWaitError(t *testing.T) {
	cfg := &config.Config{}
	r := NewClaude(cfg)

	mock := &mockCmd{waitErr: errors.New("wait error")}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	err := r.Run(context.Background(), "test", nil)
	if err == nil {
		t.Error("Run() should return error on wait failure")
	}
	if !strings.Contains(err.Error(), "Claude Code with model") && !strings.Contains(err.Error(), "failed") {
		t.Errorf("Expected Claude Code failed error, got %v", err)
	}
}

func TestClaudeOutputLineVerboseField(t *testing.T) {
	line := OutputLine{
		Text:    "[DEBUG] test output",
		IsErr:   false,
		Verbose: true,
	}

	if !line.Verbose {
		t.Error("Verbose = false, want true")
	}
}

func TestClaudeRunOutputTimestamps(t *testing.T) {
	cfg := &config.Config{Model: "claude-code/sonnet"}
	r := NewClaude(cfg)

	mock := &mockCmd{stdout: "test output line", stderr: ""}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	outputCh := make(chan OutputLine, 100)
	err := r.Run(context.Background(), "test prompt", outputCh)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	close(outputCh)
	var lines []OutputLine
	for line := range outputCh {
		lines = append(lines, line)
	}

	for i, line := range lines {
		if line.Time.IsZero() {
			t.Errorf("Line %d has zero timestamp", i)
		}
		if time.Since(line.Time) > time.Second*10 {
			t.Errorf("Line %d timestamp is too old: %v", i, line.Time)
		}
	}
}

func TestParseClaudeStreamJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantLen     int
		wantText    string
		wantVerbose bool
		wantErr     bool
	}{
		{
			name:        "system init event",
			input:       `{"type":"system","subtype":"init"}`,
			wantLen:     1,
			wantText:    "Claude initialized",
			wantVerbose: true,
		},
		{
			name:        "assistant text",
			input:       `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello world"}]}}`,
			wantLen:     1,
			wantText:    "Hello world",
			wantVerbose: false,
		},
		{
			name:        "assistant tool use",
			input:       `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read"}]}}`,
			wantLen:     1,
			wantText:    "Using tool: Read",
			wantVerbose: false,
		},
		{
			name:        "user tool result",
			input:       `{"type":"user"}`,
			wantLen:     1,
			wantText:    "Tool completed",
			wantVerbose: true,
		},
		{
			name:        "result success",
			input:       `{"type":"result","subtype":"success"}`,
			wantLen:     1,
			wantText:    "Task completed successfully",
			wantVerbose: true,
		},
		{
			name:     "result error",
			input:    `{"type":"result","subtype":"error"}`,
			wantLen:  1,
			wantText: "Task failed",
			wantErr:  true,
		},
		{
			name:        "invalid JSON returns raw line",
			input:       `not valid json`,
			wantLen:     1,
			wantText:    "not valid json",
			wantVerbose: true,
		},
		{
			name:    "empty text content ignored",
			input:   `{"type":"assistant","message":{"content":[{"type":"text","text":""}]}}`,
			wantLen: 0,
		},
		{
			name:    "multiple content items",
			input:   `{"type":"assistant","message":{"content":[{"type":"text","text":"First"},{"type":"tool_use","name":"Edit"},{"type":"text","text":"Second"}]}}`,
			wantLen: 3,
		},
		{
			name:    "unknown type returns empty",
			input:   `{"type":"unknown"}`,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputs := parseClaudeStreamJSON(tt.input)

			if len(outputs) != tt.wantLen {
				t.Errorf("parseClaudeStreamJSON() returned %d outputs, want %d", len(outputs), tt.wantLen)
				return
			}

			if tt.wantLen > 0 && tt.wantText != "" {
				if outputs[0].Text != tt.wantText {
					t.Errorf("Text = %q, want %q", outputs[0].Text, tt.wantText)
				}
				if outputs[0].Verbose != tt.wantVerbose {
					t.Errorf("Verbose = %v, want %v", outputs[0].Verbose, tt.wantVerbose)
				}
				if outputs[0].IsErr != tt.wantErr {
					t.Errorf("IsErr = %v, want %v", outputs[0].IsErr, tt.wantErr)
				}
			}
		})
	}
}

func TestParseClaudeStreamJSONTimestamps(t *testing.T) {
	before := time.Now()
	outputs := parseClaudeStreamJSON(`{"type":"assistant","message":{"content":[{"type":"text","text":"test"}]}}`)
	after := time.Now()

	if len(outputs) != 1 {
		t.Fatalf("Expected 1 output, got %d", len(outputs))
	}

	if outputs[0].Time.Before(before) || outputs[0].Time.After(after) {
		t.Errorf("Timestamp %v not between %v and %v", outputs[0].Time, before, after)
	}
}

func TestClaudeRunnerIsInternalLog(t *testing.T) {
	cfg := &config.Config{Model: "claude-code/sonnet"}
	r := NewClaude(cfg)

	tests := []struct {
		name string
		line string
		want bool
	}{
		{
			name: "debug info - internal log",
			line: "loading configuration",
			want: true,
		},
		{
			name: "status info - internal log",
			line: "checking permissions",
			want: true,
		},
		{
			name: "error message - user facing",
			line: "Error: file not found",
			want: false,
		},
		{
			name: "failed message - user facing",
			line: "Failed to connect to server",
			want: false,
		},
		{
			name: "cannot message - user facing",
			line: "Cannot access file",
			want: false,
		},
		{
			name: "unable message - user facing",
			line: "Unable to create directory",
			want: false,
		},
		{
			name: "permission denied - user facing",
			line: "Permission denied",
			want: false,
		},
		{
			name: "invalid input - user facing",
			line: "Invalid input format",
			want: false,
		},
		{
			name: "generic debug - internal log",
			line: "[DEBUG] processing request",
			want: true,
		},
		{
			name: "empty line - internal log",
			line: "",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.IsInternalLog(tt.line)
			if got != tt.want {
				t.Errorf("ClaudeRunner.IsInternalLog(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}
