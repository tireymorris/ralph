package runner

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"ralph/internal/shared/config"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{Runner: "opencode"}
	r := New(cfg)

	if r == nil {
		t.Fatal("New() returned nil")
	}

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
	stdin     io.Reader
}

func (m *mockCmd) setStdin(r io.Reader) {
	m.stdin = r
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
	cfg := &config.Config{Runner: "opencode"}
	r := newTestRunner(t, cfg)

	mock := &mockCmd{stdout: "output line", stderr: ""}
	r.CmdFunc = stubCmdFunc(mock, nil, nil)

	err := r.Run(context.Background(), "test prompt", nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunWithOutputChannel(t *testing.T) {
	cfg := &config.Config{Runner: "opencode"}
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

func TestOpenCodeDoesNotPassModelSelectionArgs(t *testing.T) {
	cfg := &config.Config{Runner: "opencode"}
	r := newTestRunner(t, cfg)

	var capturedArgs []string
	mock := &mockCmd{stdout: "ok", stderr: ""}
	r.CmdFunc = stubCmdFunc(mock, nil, &capturedArgs)

	if err := r.Run(context.Background(), "test", nil); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	assertNoModelSelectionArgs(t, capturedArgs)
}

func TestOpenCodeRunArgs(t *testing.T) {
	cfg := &config.Config{Runner: "opencode"}
	r := newTestRunner(t, cfg)

	var capturedArgs []string
	mock := &mockCmd{stdout: "ok", stderr: ""}
	r.CmdFunc = stubCmdFunc(mock, nil, &capturedArgs)

	if err := r.Run(context.Background(), "test prompt", nil); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	expectedArgs := []string{"run", "--print-logs"}
	assertArgsEqual(t, capturedArgs, expectedArgs)
	assertPromptDeliveredViaStdin(t, mock, "test prompt")
}

func TestOpenCodeSupportsLargePrompts(t *testing.T) {
	cfg := &config.Config{Runner: "opencode"}
	r := newTestRunner(t, cfg)

	prompt := strings.Repeat("implement feature ", 40000)
	mock := &mockCmd{stdout: "ok", stderr: ""}
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
			line: "INFO 2026-01-19T22:45:58 service=provider runner=test",
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

func TestNewReturnsCopilotRunner(t *testing.T) {
	cfg := &config.Config{Runner: "copilot"}
	r := New(cfg)

	if r == nil {
		t.Fatal("New() returned nil")
	}

	copilotRunner, ok := r.(*CopilotRunner)
	if !ok {
		t.Fatalf("New() returned %T, want *CopilotRunner", r)
	}

	if copilotRunner.RunnerName() != "copilot" {
		t.Errorf("RunnerName() = %q, want %q", copilotRunner.RunnerName(), "copilot")
	}

	if copilotRunner.CommandName() != "copilot" {
		t.Errorf("CommandName() = %q, want %q", copilotRunner.CommandName(), "copilot")
	}
}

func TestNewReturnsCursorAgentRunner(t *testing.T) {
	cfg := &config.Config{Runner: "cursor"}
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
	cfg := &config.Config{Runner: "claude"}
	runner := New(cfg)

	if runner == nil {
		t.Fatal("New() returned nil")
	}

	claudeRunner, ok := runner.(*ClaudeRunner)
	if !ok {
		t.Errorf("New() returned %T, want *ClaudeRunner", runner)
	}

	if claudeRunner.cfg != cfg {
		t.Error("ClaudeRunner config not set correctly")
	}
}

func TestNewReturnsOpenCodeRunner(t *testing.T) {
	cfg := &config.Config{Runner: "opencode"}
	runner := New(cfg)

	if runner == nil {
		t.Fatal("New() returned nil")
	}

	openCodeRunner, ok := runner.(*Runner)
	if !ok {
		t.Errorf("New() returned %T, want *Runner", runner)
	}

	if openCodeRunner.cfg != cfg {
		t.Error("Runner config not set correctly")
	}
}

func TestNewWithDefaultRunner(t *testing.T) {
	cfg := &config.Config{Runner: config.DefaultRunner}
	runner := New(cfg)

	if runner == nil {
		t.Fatal("New() returned nil")
	}

	_, ok := runner.(*ClaudeRunner)
	if !ok {
		t.Errorf("New() returned %T, want *ClaudeRunner for default runner", runner)
	}
}

func TestNewWithError(t *testing.T) {
	tests := []struct {
		name    string
		runner  string
		want    any
		wantErr bool
	}{
		{name: "claude", runner: "claude", want: &ClaudeRunner{}},
		{name: "open code", runner: "opencode", want: &Runner{}},
		{name: "pi", runner: "pi", want: &PiRunner{}},
		{name: "invalid", runner: "invalid-runner", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, err := NewWithError(&config.Config{Runner: tt.runner})
			if tt.wantErr {
				if err == nil {
					t.Fatal("NewWithError() should return error")
				}
				if runner != nil {
					t.Fatal("NewWithError() should return nil runner")
				}
				if !strings.Contains(err.Error(), "invalid runner configuration") {
					t.Fatalf("error = %v, want invalid runner configuration", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewWithError() error = %v", err)
			}
			if runner == nil {
				t.Fatal("NewWithError() returned nil runner")
			}
			if tt.runner == "claude" {
				_ = assertRunnerIs[*ClaudeRunner](t, runner)
			} else if tt.runner == "opencode" {
				_ = assertRunnerIs[*Runner](t, runner)
			} else {
				_ = assertRunnerIs[*PiRunner](t, runner)
			}
		})
	}
}

func TestRunnerSwitchingBetweenRuns(t *testing.T) {
	tests := []struct {
		name   string
		runner string
		want   any
	}{
		{name: "claude", runner: "claude", want: &ClaudeRunner{}},
		{name: "opencode", runner: "opencode", want: &Runner{}},
		{name: "claude-again", runner: "claude", want: &ClaudeRunner{}},
		{name: "pi", runner: "pi", want: &PiRunner{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(&config.Config{Runner: tt.runner})
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

func TestIntegrationClaudeRunnerExecution(t *testing.T) {
	cfg := &config.Config{Runner: "claude"}
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
		t.Errorf("Expected command 'claude', got %q", capturedName)
	}

	expectedArgs := []string{"--print", "--verbose", "--output-format", "stream-json", "--dangerously-skip-permissions"}
	assertArgsEqual(t, capturedArgs, expectedArgs)
}

func TestIntegrationOpenCodeRunnerExecution(t *testing.T) {
	cfg := &config.Config{Runner: "opencode"}
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

	expectedArgs := []string{"run", "--print-logs"}
	assertArgsEqual(t, capturedArgs, expectedArgs)
}

func TestIntegrationPiRunnerExecution(t *testing.T) {
	cfg := &config.Config{Runner: "pi"}
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

	expectedArgs := []string{"--print", "--mode", "json", "--no-session"}
	assertArgsEqual(t, capturedArgs, expectedArgs)
}

func TestRunnerInterfaceIsInternalLog(t *testing.T) {
	tests := []struct {
		name string
		r    RunnerInterface
		line string
		want bool
	}{
		{name: "open code internal", r: New(&config.Config{Runner: "opencode"}), line: "service=bus starting", want: true},
		{name: "open code normal", r: New(&config.Config{Runner: "opencode"}), line: "Regular output", want: false},
		{name: "claude internal", r: New(&config.Config{Runner: "claude"}), line: "debug info", want: true},
		{name: "claude user error", r: New(&config.Config{Runner: "claude"}), line: "Error: file not found", want: false},
		{name: "pi internal", r: New(&config.Config{Runner: "pi"}), line: "debug info", want: true},
		{name: "pi user error", r: New(&config.Config{Runner: "pi"}), line: "Error: file not found", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.IsInternalLog(tt.line); got != tt.want {
				t.Errorf("IsInternalLog(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}
