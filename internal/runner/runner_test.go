package runner

import (
	"context"
	"errors"
	"io"
	"os/exec"
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
	if r.cfg != cfg {
		t.Error("New() did not set config correctly")
	}
	if r.CmdFunc == nil {
		t.Error("CmdFunc should not be nil")
	}
}

type mockWriteCloser struct {
	io.Writer
	closed bool
}

func (m *mockWriteCloser) Close() error {
	m.closed = true
	return nil
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
	stdinErr  error
	stdoutErr error
	stderrErr error
	startErr  error
	waitErr   error
	stdout    string
	stderr    string
}

func (m *mockCmd) StdinPipe() (io.WriteCloser, error) {
	if m.stdinErr != nil {
		return nil, m.stdinErr
	}
	return &mockWriteCloser{Writer: &strings.Builder{}}, nil
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

func TestRunOpenCodeSuccess(t *testing.T) {
	cfg := &config.Config{Model: "test-model"}
	r := New(cfg)

	mock := &mockCmd{stdout: "output line", stderr: ""}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	result, err := r.RunOpenCode(context.Background(), "test prompt", nil)
	if err != nil {
		t.Fatalf("RunOpenCode() error = %v", err)
	}
	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestRunOpenCodeWithOutputChannel(t *testing.T) {
	cfg := &config.Config{Model: "test-model"}
	r := New(cfg)

	mock := &mockCmd{stdout: "line1\nline2", stderr: "err1"}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	outputCh := make(chan OutputLine, 100)
	result, err := r.RunOpenCode(context.Background(), "test", outputCh)
	if err != nil {
		t.Fatalf("RunOpenCode() error = %v", err)
	}
	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

// Note: TestRunOpenCodeStdinError was removed because RunOpenCode
// passes the prompt as a command-line argument, not via stdin.

func TestRunOpenCodeStdoutError(t *testing.T) {
	cfg := &config.Config{}
	r := New(cfg)

	mock := &mockCmd{stdoutErr: errors.New("stdout error")}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	_, err := r.RunOpenCode(context.Background(), "test", nil)
	if err == nil {
		t.Error("RunOpenCode() should error on stdout failure")
	}
}

func TestRunOpenCodeStderrError(t *testing.T) {
	cfg := &config.Config{}
	r := New(cfg)

	mock := &mockCmd{stderrErr: errors.New("stderr error")}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	_, err := r.RunOpenCode(context.Background(), "test", nil)
	if err == nil {
		t.Error("RunOpenCode() should error on stderr failure")
	}
}

func TestRunOpenCodeStartError(t *testing.T) {
	cfg := &config.Config{}
	r := New(cfg)

	mock := &mockCmd{startErr: errors.New("start error")}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	_, err := r.RunOpenCode(context.Background(), "test", nil)
	if err == nil {
		t.Error("RunOpenCode() should error on start failure")
	}
}

func TestRunOpenCodeWaitError(t *testing.T) {
	cfg := &config.Config{}
	r := New(cfg)

	mock := &mockCmd{waitErr: errors.New("wait error")}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	result, err := r.RunOpenCode(context.Background(), "test", nil)
	if err != nil {
		t.Fatalf("RunOpenCode() should not return error, got %v", err)
	}
	if result.Error == nil {
		t.Error("Result.Error should be set")
	}
}

type exitError struct {
	code int
}

func (e *exitError) Error() string {
	return "exit error"
}

func (e *exitError) ExitCode() int {
	return e.code
}

func TestRunOpenCodeExitError(t *testing.T) {
	cfg := &config.Config{}
	r := New(cfg)

	mock := &mockCmd{waitErr: &exec.ExitError{}}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		return mock
	}

	result, err := r.RunOpenCode(context.Background(), "test", nil)
	if err != nil {
		t.Fatalf("RunOpenCode() should not return error, got %v", err)
	}
	if result.Error != nil {
		t.Error("Result.Error should be nil for exit error")
	}
}

func TestRunOpenCodeNoModel(t *testing.T) {
	cfg := &config.Config{Model: ""}
	r := New(cfg)

	var capturedArgs []string
	mock := &mockCmd{}
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		capturedArgs = args
		return mock
	}

	r.RunOpenCode(context.Background(), "test", nil)

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

func TestRealCmdMethods(t *testing.T) {
	rc := &realCmd{Cmd: nil}
	if rc == nil {
		t.Error("realCmd should not be nil")
	}
}

func TestRealCmdPipes(t *testing.T) {
	cmdFunc := defaultCmdFunc("")
	cmd := cmdFunc(context.Background(), "echo", "test")
	rc := cmd.(*realCmd)

	stdin, err := rc.StdinPipe()
	if err != nil {
		t.Errorf("StdinPipe() error = %v", err)
	}
	if stdin != nil {
		stdin.Close()
	}

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

func TestCleanOutput(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "no escape codes",
			input:  "plain text",
			expect: "plain text",
		},
		{
			name:   "simple escape code",
			input:  "\x1b[31mred\x1b[0m",
			expect: "red",
		},
		{
			name:   "multiple escape codes",
			input:  "\x1b[1m\x1b[32mbold green\x1b[0m normal",
			expect: "bold green normal",
		},
		{
			name:   "cursor movement",
			input:  "\x1b[2Kmoved\x1b[1A",
			expect: "moved",
		},
		{
			name:   "whitespace trimming",
			input:  "  \n  trimmed  \n  ",
			expect: "trimmed",
		},
		{
			name:   "complex escape sequence",
			input:  "\x1b[38;5;196mcolored\x1b[0m",
			expect: "colored",
		},
		{
			name:   "empty string",
			input:  "",
			expect: "",
		},
		{
			name:   "only escape codes",
			input:  "\x1b[0m\x1b[1m\x1b[32m",
			expect: "",
		},
		{
			name:   "escape at end",
			input:  "text\x1b[0m",
			expect: "text",
		},
		{
			name:   "incomplete escape sequence",
			input:  "text\x1b[",
			expect: "text",
		},
		{
			name:   "OSC sequence with BEL terminator",
			input:  "\x1b]0;window title\x07text here",
			expect: "text here",
		},
		{
			name:   "OSC sequence with ST terminator",
			input:  "\x1b]0;window title\x1b\\text here",
			expect: "text here",
		},
		{
			name:   "mixed CSI and OSC sequences",
			input:  "\x1b]0;title\x07\x1b[32mgreen\x1b[0m",
			expect: "green",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanOutput(tt.input)
			if got != tt.expect {
				t.Errorf("CleanOutput() = %q, want %q", got, tt.expect)
			}
		})
	}
}

func TestIsCSITerminator(t *testing.T) {
	tests := []struct {
		input byte
		want  bool
	}{
		{'A', true},
		{'Z', true},
		{'a', true},
		{'z', true},
		{'m', true},
		{'K', true},
		{'0', false},
		{'9', false},
		{';', false},
		{'[', false},
		{' ', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			got := isCSITerminator(tt.input)
			if got != tt.want {
				t.Errorf("isCSITerminator(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
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

func TestResult(t *testing.T) {
	result := Result{
		Output:   "output text",
		ExitCode: 1,
		Error:    nil,
	}

	if result.Output != "output text" {
		t.Errorf("Output = %q, want %q", result.Output, "output text")
	}
	if result.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", result.ExitCode)
	}
	if result.Error != nil {
		t.Errorf("Error = %v, want nil", result.Error)
	}
}

func TestIsVerboseLogLine(t *testing.T) {
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
			name: "tool call output - not verbose",
			line: "Running tool: read_file",
			want: false,
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
			name: "COMPLETED marker - not verbose",
			line: "COMPLETED: done",
			want: false,
		},
		{
			name: "empty line - not verbose",
			line: "",
			want: false,
		},
		{
			name: "short line - not verbose",
			line: "OK",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isVerboseLogLine(tt.line)
			if got != tt.want {
				t.Errorf("isVerboseLogLine(%q) = %v, want %v", tt.line, got, tt.want)
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
