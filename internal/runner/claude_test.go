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
	if !strings.Contains(err.Error(), "claude failed") {
		t.Errorf("Expected claude failed error, got %v", err)
	}
}

func TestIsClaudeVerboseLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{
			name: "debug log",
			line: "[DEBUG] Initializing session",
			want: true,
		},
		{
			name: "info log",
			line: "[INFO] Model loaded",
			want: true,
		},
		{
			name: "warn log",
			line: "[WARN] Rate limit approaching",
			want: true,
		},
		{
			name: "error log",
			line: "[ERROR] Connection failed",
			want: true,
		},
		{
			name: "trace log",
			line: "TRACE: Processing request",
			want: true,
		},
		{
			name: "tool execution",
			line: "Tool execution: read_file",
			want: true,
		},
		{
			name: "API request",
			line: "API request: POST /v1/messages",
			want: true,
		},
		{
			name: "API response",
			line: "API response: 200 OK",
			want: true,
		},
		{
			name: "token usage",
			line: "Token usage: 1234 input, 567 output",
			want: true,
		},
		{
			name: "process ID",
			line: "Process ID: 12345",
			want: true,
		},
		{
			name: "working directory",
			line: "Working directory: /Users/test/project",
			want: true,
		},
		{
			name: "git repository",
			line: "Git repository: /Users/test/project",
			want: true,
		},
		{
			name: "session ID",
			line: "Session ID: sess_1234567890",
			want: true,
		},
		{
			name: "model info",
			line: "Model: claude-3.5-sonnet",
			want: true,
		},
		{
			name: "temperature",
			line: "Temperature: 0.7",
			want: true,
		},
		{
			name: "max tokens",
			line: "Max tokens: 4096",
			want: true,
		},
		{
			name: "duration",
			line: "duration=1.234s",
			want: true,
		},
		{
			name: "status",
			line: "status=completed",
			want: true,
		},
		{
			name: "bytes",
			line: "bytes=1024",
			want: true,
		},
		{
			name: "files",
			line: "files=5",
			want: true,
		},
		{
			name: "request ID",
			line: "request_id=req_1234567890",
			want: true,
		},
		{
			name: "timestamp",
			line: "timestamp=2023-01-01T00:00:00Z",
			want: true,
		},
		{
			name: "tree line with pipe",
			line: " ├─ Reading file",
			want: true,
		},
		{
			name: "tree line with corner",
			line: " └─ Complete",
			want: true,
		},
		{
			name: "regular output - not verbose",
			line: "Implementing feature...",
			want: false,
		},
		{
			name: "error output - not verbose",
			line: "Error: something went wrong",
			want: false,
		},
		{
			name: "empty line - not verbose",
			line: "",
			want: false,
		},
		{
			name: "simple output - not verbose",
			line: "Hello, world!",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isClaudeVerboseLine(tt.line)
			if got != tt.want {
				t.Errorf("isClaudeVerboseLine(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
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
