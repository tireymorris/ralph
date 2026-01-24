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
	if r.cfg != cfg {
		t.Error("New() did not set config correctly")
	}
	if r.CmdFunc == nil {
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
	r := New(cfg)

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
	r := New(cfg)

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
	r := New(cfg)

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
	r := New(cfg)

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
	r := New(cfg)

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
	r := New(cfg)

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
	r := New(cfg)

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

func TestIsVerboseLine(t *testing.T) {
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
			got := isVerboseLine(tt.line)
			if got != tt.want {
				t.Errorf("isVerboseLine(%q) = %v, want %v", tt.line, got, tt.want)
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
