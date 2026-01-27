package runner

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"ralph/internal/config"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{Model: "test-model"}
	r := New(cfg)

	if r == nil {
		t.Fatal("New() returned nil")
	}

	// Test with concrete Runner type since test-model is not a claude-code model
	runner, ok := r.(*Runner)
	if !ok {
		t.Fatalf("New() returned %T, want *Runner", r)
	}

	if runner.cfg != cfg {
		t.Error("New() did not set config correctly")
	}
	if runner.CmdFunc == nil {
		t.Error("CmdFunc should not be nil")
	}
}

type mockReadCloser struct {
	*strings.Reader
	closed bool
}

func (m *mockReadCloser) Close() error {
	m.closed = true
	return nil
}

type mockCmd struct {
	stdoutErr error
	stderrErr error
	startErr  error
	waitErr   error
	stdout    string
	stderr    string
}

func (m *mockCmd) StdoutPipe() (io.ReadCloser, error) {
	if m.stdoutErr != nil {
		return nil, m.stdoutErr
	}
	return &mockReadCloser{Reader: strings.NewReader(m.stdout)}, nil
}

func (m *mockCmd) StderrPipe() (io.ReadCloser, error) {
	if m.stderrErr != nil {
		return nil, m.stderrErr
	}
	return &mockReadCloser{Reader: strings.NewReader(m.stderr)}, nil
}

func (m *mockCmd) Start() error {
	return m.startErr
}

func (m *mockCmd) Wait() error {
	return m.waitErr
}

func TestRunSuccess(t *testing.T) {
	cfg := &config.Config{Model: "test-model"}
	r := &Runner{cfg: cfg, CmdFunc: defaultCmdFunc(cfg.WorkDir)}

	mock := &mockCmd{stdout: "output line", stderr: ""}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	err := r.Run(context.Background(), "test prompt", nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunWithOutputChannel(t *testing.T) {
	cfg := &config.Config{Model: "test-model"}
	r := &Runner{cfg: cfg, CmdFunc: defaultCmdFunc(cfg.WorkDir)}

	mock := &mockCmd{stdout: "line1\nline2", stderr: "err1"}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	outputCh := make(chan OutputLine, 100)
	err := r.Run(context.Background(), "test", outputCh)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunStdoutError(t *testing.T) {
	cfg := &config.Config{}
	r := &Runner{cfg: cfg, CmdFunc: defaultCmdFunc(cfg.WorkDir)}

	mock := &mockCmd{stdoutErr: errors.New("stdout error")}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	err := r.Run(context.Background(), "test", nil)
	if err == nil {
		t.Error("Run() should error on stdout failure")
	}
}

func TestRunStderrError(t *testing.T) {
	cfg := &config.Config{}
	r := &Runner{cfg: cfg, CmdFunc: defaultCmdFunc(cfg.WorkDir)}

	mock := &mockCmd{stderrErr: errors.New("stderr error")}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	err := r.Run(context.Background(), "test", nil)
	if err == nil {
		t.Error("Run() should error on stderr failure")
	}
}

func TestRunStartError(t *testing.T) {
	cfg := &config.Config{}
	r := &Runner{cfg: cfg, CmdFunc: defaultCmdFunc(cfg.WorkDir)}

	mock := &mockCmd{startErr: errors.New("start error")}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	err := r.Run(context.Background(), "test", nil)
	if err == nil {
		t.Error("Run() should error on start failure")
	}
}

func TestRunWaitError(t *testing.T) {
	cfg := &config.Config{}
	r := &Runner{cfg: cfg, CmdFunc: defaultCmdFunc(cfg.WorkDir)}

	mock := &mockCmd{waitErr: errors.New("wait error")}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	err := r.Run(context.Background(), "test", nil)
	if err == nil {
		t.Error("Run() should return error on wait failure")
	}
}

func TestRunNoModel(t *testing.T) {
	cfg := &config.Config{Model: ""}
	r := &Runner{cfg: cfg, CmdFunc: defaultCmdFunc(cfg.WorkDir)}

	var capturedArgs []string
	mock := &mockCmd{}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		capturedArgs = args
		return mock
	}

	r.Run(context.Background(), "test", nil)

	for _, arg := range capturedArgs {
		if arg == "--model" {
			t.Error("--model should not be in args when Model is empty")
		}
	}
}

func TestDefaultCmdFunc(t *testing.T) {
	cmdFunc := defaultCmdFunc("")
	cmd := cmdFunc(context.Background(), "echo", "test")
	if cmd == nil {
		t.Error("defaultCmdFunc() returned nil")
	}
}

func TestDefaultCmdFuncWithWorkDir(t *testing.T) {
	tmpDir := t.TempDir()
	cmdFunc := defaultCmdFunc(tmpDir)
	cmd := cmdFunc(context.Background(), "pwd")
	if cmd == nil {
		t.Error("defaultCmdFunc() returned nil")
	}
	rc := cmd.(*realCmd)
	if rc.Cmd.Dir != tmpDir {
		t.Errorf("Cmd.Dir = %q, want %q", rc.Cmd.Dir, tmpDir)
	}
}

func TestRealCmdPipes(t *testing.T) {
	cmdFunc := defaultCmdFunc("")
	cmd := cmdFunc(context.Background(), "echo", "test")
	rc := cmd.(*realCmd)

	stdout, err := rc.StdoutPipe()
	if err != nil {
		t.Errorf("StdoutPipe() error = %v", err)
	}
	if stdout != nil {
		stdout.Close()
	}

	stderr, err := rc.StderrPipe()
	if err != nil {
		t.Errorf("StderrPipe() error = %v", err)
	}
	if stderr != nil {
		stderr.Close()
	}
}

func TestRealCmdStartWait(t *testing.T) {
	cmdFunc := defaultCmdFunc("")
	cmd := cmdFunc(context.Background(), "echo", "test")
	rc := cmd.(*realCmd)

	err := rc.Start()
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}

	err = rc.Wait()
	if err != nil {
		t.Errorf("Wait() error = %v", err)
	}
}

func TestOutputLine(t *testing.T) {
	line := OutputLine{
		Text:  "test output",
		IsErr: true,
	}

	if line.Text != "test output" {
		t.Errorf("Text = %q, want %q", line.Text, "test output")
	}
	if !line.IsErr {
		t.Error("IsErr = false, want true")
	}
}

func TestOpenCodeInternalLogDetection(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{
			name: "service bus log",
			line: "INFO 2026-01-19T22:45:58 +22ms service=bus type=message.part.updated publishing",
			want: true,
		},
		{
			name: "debug log",
			line: "DEBUG 2026-01-19T22:45:58 +0ms service=provider starting",
			want: true,
		},
		{
			name: "warn log",
			line: "WARN 2026-01-19T22:45:58 +0ms service=session warning",
			want: true,
		},
		{
			name: "error log with timestamp",
			line: "ERROR 2026-01-19T22:51:36 +1ms service=default e=Out of memory rejection",
			want: true,
		},
		{
			name: "service provider log",
			line: "INFO 2026-01-19T22:45:58 service=provider model=test",
			want: true,
		},
		{
			name: "service lsp log",
			line: "INFO 2026-01-19T22:45:58 service=lsp initializing",
			want: true,
		},
		{
			name: "git tracking line",
			line: " cwd=/Users/tmorris/workspace/gba git=/Users/tmorris/.local/share/opencode/snapshot/608d3c tracking",
			want: true,
		},
		{
			name: "cwd status line",
			line: " cwd=/Users/tmorris/workspace/gba something",
			want: true,
		},
		{
			name: "stderr prefix line",
			line: " stderr=Saved lockfile",
			want: true,
		},
		{
			name: "package check line",
			line: "Checked 3 installs across 4 packages (no changes) [2.00ms]",
			want: true,
		},
		{
			name: "package install line",
			line: "installed @opencode-ai/plugin@1.1.25",
			want: true,
		},
		{
			name: "tool call output - not internal log",
			line: "Running tool: read_file",
			want: false,
		},
		{
			name: "regular output - not internal log",
			line: "Implementing feature...",
			want: false,
		},
		{
			name: "error output - not internal log",
			line: "Error: something went wrong",
			want: false,
		},
		{
			name: "empty line - not internal log",
			line: "",
			want: false,
		},
		{
			name: "short line - not internal log",
			line: "OK",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isOpenCodeInternalLog(tt.line)
			if got != tt.want {
				t.Errorf("isOpenCodeInternalLog(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestOutputLineVerboseField(t *testing.T) {
	line := OutputLine{
		Text:    "test output",
		IsErr:   false,
		Verbose: true,
	}

	if !line.Verbose {
		t.Error("Verbose = false, want true")
	}
}

func TestIsClaudeCodeModel(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  bool
	}{
		{
			name:  "sonnet",
			model: "claude-code/sonnet",
			want:  true,
		},
		{
			name:  "haiku",
			model: "claude-code/haiku",
			want:  true,
		},
		{
			name:  "claude-3-opus",
			model: "claude-code/claude-3-opus",
			want:  true,
		},
		{
			name:  "opencode big-pickle",
			model: "opencode/big-pickle",
			want:  false,
		},
		{
			name:  "opencode big-pickle",
			model: "opencode/big-pickle",
			want:  false,
		},
		{
			name:  "empty model",
			model: "",
			want:  false,
		},
		{
			name:  "partial claude prefix",
			model: "claude",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isClaudeCodeModel(tt.model)
			if got != tt.want {
				t.Errorf("isClaudeCodeModel(%q) = %v, want %v", tt.model, got, tt.want)
			}
		})
	}
}

func TestNewReturnsClaudeRunner(t *testing.T) {
	cfg := &config.Config{Model: "claude-code/sonnet"}
	runner := New(cfg)

	if runner == nil {
		t.Fatal("New() returned nil")
	}

	// Check that we got a ClaudeRunner
	claudeRunner, ok := runner.(*ClaudeRunner)
	if !ok {
		t.Errorf("New() returned %T, want *ClaudeRunner", runner)
	}

	if claudeRunner.cfg != cfg {
		t.Error("ClaudeRunner config not set correctly")
	}
}

func TestNewReturnsOpenCodeRunner(t *testing.T) {
	cfg := &config.Config{Model: "opencode/big-pickle"}
	runner := New(cfg)

	if runner == nil {
		t.Fatal("New() returned nil")
	}

	// Check that we got an OpenCode Runner
	openCodeRunner, ok := runner.(*Runner)
	if !ok {
		t.Errorf("New() returned %T, want *Runner", runner)
	}

	if openCodeRunner.cfg != cfg {
		t.Error("Runner config not set correctly")
	}
}

func TestNewWithDefaultModel(t *testing.T) {
	cfg := &config.Config{Model: config.DefaultModel}
	runner := New(cfg)

	if runner == nil {
		t.Fatal("New() returned nil")
	}

	// Should return OpenCode Runner for default model (opencode/big-pickle)
	_, ok := runner.(*Runner)
	if !ok {
		t.Errorf("New() returned %T, want *Runner for default model", runner)
	}
}

func TestNewWithErrorValidClaudeModel(t *testing.T) {
	cfg := &config.Config{Model: "claude-code/sonnet"}
	runner, err := NewWithError(cfg)

	if err != nil {
		t.Fatalf("NewWithError() error = %v", err)
	}

	if runner == nil {
		t.Fatal("NewWithError() returned nil runner")
	}

	_, ok := runner.(*ClaudeRunner)
	if !ok {
		t.Errorf("NewWithError() returned %T, want *ClaudeRunner", runner)
	}
}

func TestNewWithErrorValidOpenCodeModel(t *testing.T) {
	cfg := &config.Config{Model: "opencode/big-pickle"}
	runner, err := NewWithError(cfg)

	if err != nil {
		t.Fatalf("NewWithError() error = %v", err)
	}

	if runner == nil {
		t.Fatal("NewWithError() returned nil runner")
	}

	_, ok := runner.(*Runner)
	if !ok {
		t.Errorf("NewWithError() returned %T, want *Runner", runner)
	}
}

func TestNewWithErrorInvalidModel(t *testing.T) {
	cfg := &config.Config{Model: "invalid-model"}
	runner, err := NewWithError(cfg)

	if err == nil {
		t.Error("NewWithError() should return error for invalid model")
	}

	if runner != nil {
		t.Error("NewWithError() should return nil runner for invalid model")
	}

	expectedMsg := "invalid model configuration"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Error message = %v, want to contain %v", err.Error(), expectedMsg)
	}
}

func TestModelSwitchingBetweenRuns(t *testing.T) {
	// Test Claude Code model
	claudeCfg := &config.Config{Model: "claude-code/sonnet"}
	runner1 := New(claudeCfg)

	_, ok1 := runner1.(*ClaudeRunner)
	if !ok1 {
		t.Errorf("First New() call returned %T, want *ClaudeRunner", runner1)
	}

	// Test OpenCode model in second run
	openCodeCfg := &config.Config{Model: "opencode/big-pickle"}
	runner2 := New(openCodeCfg)

	_, ok2 := runner2.(*Runner)
	if !ok2 {
		t.Errorf("Second New() call returned %T, want *Runner", runner2)
	}

	// Test switching back to Claude Code
	runner3 := New(claudeCfg)
	_, ok3 := runner3.(*ClaudeRunner)
	if !ok3 {
		t.Errorf("Third New() call returned %T, want *ClaudeRunner", runner3)
	}
}

func TestIntegrationClaudeModelExecution(t *testing.T) {
	cfg := &config.Config{Model: "claude-code/sonnet"}
	runner := New(cfg)

	// Mock the command execution for Claude runner
	var capturedName string
	var capturedArgs []string
	mock := &mockCmd{stdout: "claude output", stderr: ""}

	// Type assert to ClaudeRunner to set the mock
	claudeRunner, ok := runner.(*ClaudeRunner)
	if !ok {
		t.Fatalf("Expected *ClaudeRunner, got %T", runner)
	}

	claudeRunner.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		capturedName = name
		capturedArgs = args
		return mock
	}

	err := runner.Run(context.Background(), "test prompt", nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if capturedName != "claude" {
		t.Errorf("Expected command 'claude-code', got %q", capturedName)
	}

	// Verify Claude-specific arguments
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

func TestIntegrationOpenCodeModelExecution(t *testing.T) {
	cfg := &config.Config{Model: "opencode/big-pickle"}
	runner := New(cfg)

	// Mock the command execution for OpenCode runner
	var capturedName string
	var capturedArgs []string
	mock := &mockCmd{stdout: "opencode output", stderr: ""}

	// Type assert to Runner to set the mock
	openCodeRunner, ok := runner.(*Runner)
	if !ok {
		t.Fatalf("Expected *Runner, got %T", runner)
	}

	openCodeRunner.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		capturedName = name
		capturedArgs = args
		return mock
	}

	err := runner.Run(context.Background(), "test prompt", nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if capturedName != "opencode" {
		t.Errorf("Expected command 'opencode', got %q", capturedName)
	}

	// Verify OpenCode-specific arguments
	expectedArgs := []string{"run", "--print-logs", "--model", "opencode/big-pickle", "test prompt"}
	if len(capturedArgs) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d", len(expectedArgs), len(capturedArgs))
	}
	for i, expected := range expectedArgs {
		if capturedArgs[i] != expected {
			t.Errorf("Arg %d: expected %q, got %q", i, expected, capturedArgs[i])
		}
	}
}

func TestRunnerInterfaceIsInternalLog(t *testing.T) {
	// Test OpenCode runner
	openCodeCfg := &config.Config{Model: "opencode/big-pickle"}
	openCodeRunner := New(openCodeCfg)

	// Test internal log detection for OpenCode
	tests := []struct {
		line string
		want bool
	}{
		{"service=bus starting", true},
		{"Regular output", false},
		{"INFO 2026-01-19T22:45:58 service=provider test", true},
	}

	for _, tt := range tests {
		got := openCodeRunner.IsInternalLog(tt.line)
		if got != tt.want {
			t.Errorf("OpenCodeRunner.IsInternalLog(%q) = %v, want %v", tt.line, got, tt.want)
		}
	}

	// Test Claude runner
	claudeCfg := &config.Config{Model: "claude-code/sonnet"}
	claudeRunner := New(claudeCfg)

	// Test internal log detection for Claude (should treat most stderr as internal)
	claudeTests := []struct {
		line string
		want bool
	}{
		{"debug info", true},
		{"Error: file not found", false}, // User-facing error
		{"Failed to load", false},        // User-facing error
		{"loading config", true},
	}

	for _, tt := range claudeTests {
		got := claudeRunner.IsInternalLog(tt.line)
		if got != tt.want {
			t.Errorf("ClaudeRunner.IsInternalLog(%q) = %v, want %v", tt.line, got, tt.want)
		}
	}
}
