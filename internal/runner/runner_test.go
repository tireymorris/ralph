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
	r := newTestRunner(t, cfg)

	mock := &mockCmd{stdout: "output line", stderr: ""}
	r.CmdFunc = stubCmdFunc(mock, nil, nil)

	err := r.Run(context.Background(), "test prompt", nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunWithOutputChannel(t *testing.T) {
	cfg := &config.Config{Model: "test-model"}
	r := newTestRunner(t, cfg)

	mock := &mockCmd{stdout: "line1\nline2", stderr: "err1"}
	r.CmdFunc = stubCmdFunc(mock, nil, nil)

	outputCh := make(chan OutputLine, 100)
	err := r.Run(context.Background(), "test", outputCh)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunStdoutError(t *testing.T) {
	cfg := &config.Config{}
	r := newTestRunner(t, cfg)

	mock := &mockCmd{stdoutErr: errors.New("stdout error")}
	r.CmdFunc = stubCmdFunc(mock, nil, nil)

	err := r.Run(context.Background(), "test", nil)
	if err == nil {
		t.Error("Run() should error on stdout failure")
	}
}

func TestRunStderrError(t *testing.T) {
	cfg := &config.Config{}
	r := newTestRunner(t, cfg)

	mock := &mockCmd{stderrErr: errors.New("stderr error")}
	r.CmdFunc = stubCmdFunc(mock, nil, nil)

	err := r.Run(context.Background(), "test", nil)
	if err == nil {
		t.Error("Run() should error on stderr failure")
	}
}

func TestRunStartError(t *testing.T) {
	cfg := &config.Config{}
	r := newTestRunner(t, cfg)

	mock := &mockCmd{startErr: errors.New("start error")}
	r.CmdFunc = stubCmdFunc(mock, nil, nil)

	err := r.Run(context.Background(), "test", nil)
	if err == nil {
		t.Error("Run() should error on start failure")
	}
}

func TestRunWaitError(t *testing.T) {
	cfg := &config.Config{}
	r := newTestRunner(t, cfg)

	mock := &mockCmd{waitErr: errors.New("wait error")}
	r.CmdFunc = stubCmdFunc(mock, nil, nil)

	err := r.Run(context.Background(), "test", nil)
	if err == nil {
		t.Error("Run() should return error on wait failure")
	}
}

func TestRunNoModel(t *testing.T) {
	cfg := &config.Config{Model: ""}
	r := newTestRunner(t, cfg)

	var capturedArgs []string
	mock := &mockCmd{}
	r.CmdFunc = stubCmdFunc(mock, nil, &capturedArgs)

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

func TestNewReturnsCursorAgentRunner(t *testing.T) {
	cfg := &config.Config{Model: "cursor-agent/sonnet-4"}
	r := New(cfg)

	if r == nil {
		t.Fatal("New() returned nil")
	}

	car, ok := r.(*CursorAgentRunner)
	if !ok {
		t.Fatalf("New() returned %T, want *CursorAgentRunner", r)
	}

	if car.RunnerName() != "cursor-agent" {
		t.Errorf("RunnerName() = %q, want %q", car.RunnerName(), "cursor-agent")
	}

	if car.CommandName() != "cursor-agent" {
		t.Errorf("CommandName() = %q, want %q", car.CommandName(), "cursor-agent")
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

	_, ok := runner.(*ClaudeRunner)
	if !ok {
		t.Errorf("New() returned %T, want *ClaudeRunner for default model", runner)
	}
}

func TestNewWithError(t *testing.T) {
	tests := []struct {
		name    string
		model   string
		want    any
		wantErr bool
	}{
		{name: "claude", model: "claude-code/sonnet", want: &ClaudeRunner{}},
		{name: "open code", model: "opencode/big-pickle", want: &Runner{}},
		{name: "pi", model: "pi/sonnet", want: &PiRunner{}},
		{name: "invalid", model: "invalid-model", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, err := NewWithError(&config.Config{Model: tt.model})
			if tt.wantErr {
				if err == nil {
					t.Fatal("NewWithError() should return error")
				}
				if runner != nil {
					t.Fatal("NewWithError() should return nil runner")
				}
				if !strings.Contains(err.Error(), "invalid model configuration") {
					t.Fatalf("error = %v, want invalid model configuration", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewWithError() error = %v", err)
			}
			if runner == nil {
				t.Fatal("NewWithError() returned nil runner")
			}
			if tt.model == "claude-code/sonnet" {
				_ = assertRunnerIs[*ClaudeRunner](t, runner)
			} else if tt.model == "opencode/big-pickle" {
				_ = assertRunnerIs[*Runner](t, runner)
			} else {
				_ = assertRunnerIs[*PiRunner](t, runner)
			}
		})
	}
}

func TestModelSwitchingBetweenRuns(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  any
	}{
		{name: "claude", model: "claude-code/sonnet", want: &ClaudeRunner{}},
		{name: "opencode", model: "opencode/big-pickle", want: &Runner{}},
		{name: "claude-again", model: "claude-code/sonnet", want: &ClaudeRunner{}},
		{name: "pi", model: "pi/sonnet", want: &PiRunner{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(&config.Config{Model: tt.model})
			if r == nil {
				t.Fatal("New() returned nil")
			}
			switch tt.want.(type) {
			case *ClaudeRunner:
				_ = assertRunnerIs[*ClaudeRunner](t, r)
			case *Runner:
				_ = assertRunnerIs[*Runner](t, r)
			case *PiRunner:
				_ = assertRunnerIs[*PiRunner](t, r)
			}
		})
	}
}

func TestIntegrationClaudeModelExecution(t *testing.T) {
	cfg := &config.Config{Model: "claude-code/sonnet"}
	runner := New(cfg)

	mock := &mockCmd{stdout: "claude output", stderr: ""}
	claudeRunner, ok := runner.(*ClaudeRunner)
	if !ok {
		t.Fatalf("Expected *ClaudeRunner, got %T", runner)
	}
	var capturedName string
	var capturedArgs []string
	claudeRunner.CmdFunc = stubCmdFunc(mock, &capturedName, &capturedArgs)

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

	mock := &mockCmd{stdout: "opencode output", stderr: ""}
	openCodeRunner, ok := runner.(*Runner)
	if !ok {
		t.Fatalf("Expected *Runner, got %T", runner)
	}
	var capturedName string
	var capturedArgs []string
	openCodeRunner.CmdFunc = stubCmdFunc(mock, &capturedName, &capturedArgs)

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

func TestIntegrationPiModelExecution(t *testing.T) {
	cfg := &config.Config{Model: "pi/sonnet"}
	runner := New(cfg)

	mock := &mockCmd{stdout: `{"type":"session","version":3}`, stderr: ""}
	piR, ok := runner.(*PiRunner)
	if !ok {
		t.Fatalf("Expected *PiRunner, got %T", runner)
	}
	var capturedName string
	var capturedArgs []string
	piR.CmdFunc = stubCmdFunc(mock, &capturedName, &capturedArgs)

	err := runner.Run(context.Background(), "test prompt", nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if capturedName != "pi" {
		t.Errorf("Expected command 'pi', got %q", capturedName)
	}

	expectedArgs := []string{"--print", "--mode", "json", "--no-session", "--model", "sonnet", "test prompt"}
	if len(capturedArgs) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d: %v", len(expectedArgs), len(capturedArgs), capturedArgs)
	}
	for i, expected := range expectedArgs {
		if capturedArgs[i] != expected {
			t.Errorf("Arg %d: expected %q, got %q", i, expected, capturedArgs[i])
		}
	}
}

func TestRunnerInterfaceIsInternalLog(t *testing.T) {
	tests := []struct {
		name string
		r    RunnerInterface
		line string
		want bool
	}{
		{name: "open code internal", r: New(&config.Config{Model: "opencode/big-pickle"}), line: "service=bus starting", want: true},
		{name: "open code normal", r: New(&config.Config{Model: "opencode/big-pickle"}), line: "Regular output", want: false},
		{name: "claude internal", r: New(&config.Config{Model: "claude-code/sonnet"}), line: "debug info", want: true},
		{name: "claude user error", r: New(&config.Config{Model: "claude-code/sonnet"}), line: "Error: file not found", want: false},
		{name: "pi internal", r: New(&config.Config{Model: "pi/sonnet"}), line: "debug info", want: true},
		{name: "pi user error", r: New(&config.Config{Model: "pi/sonnet"}), line: "Error: file not found", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.IsInternalLog(tt.line); got != tt.want {
				t.Errorf("IsInternalLog(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}
