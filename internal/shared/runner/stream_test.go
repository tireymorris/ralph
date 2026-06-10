package runner

import (
	"errors"
	"io"
	"os/exec"
	"strings"
	"testing"

	"ralph/internal/shared/constants"
)

func passthroughTransform(line string) []OutputLine {
	return []OutputLine{{Text: line}}
}

func collectPipeLines(t *testing.T, input string) []string {
	t.Helper()
	outputCh := make(chan OutputLine, 16)
	if err := readPipeLines(strings.NewReader(input), outputCh, passthroughTransform); err != nil {
		t.Fatalf("readPipeLines() error = %v", err)
	}
	close(outputCh)
	var lines []string
	for out := range outputCh {
		lines = append(lines, out.Text)
	}
	return lines
}

func TestReadPipeLinesHandlesLinesBeyondScannerLimit(t *testing.T) {
	longLine := strings.Repeat("x", 2*1024*1024)
	lines := collectPipeLines(t, "before\n"+longLine+"\nafter\n")

	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	if lines[1] != longLine {
		t.Fatalf("long line corrupted: got %d bytes, want %d", len(lines[1]), len(longLine))
	}
}

func TestReadPipeLinesTrimsCRLF(t *testing.T) {
	lines := collectPipeLines(t, "one\r\ntwo\r\n")

	if len(lines) != 2 || lines[0] != "one" || lines[1] != "two" {
		t.Fatalf("got %q, want [one two]", lines)
	}
}

func TestReadPipeLinesEmitsFinalLineWithoutNewline(t *testing.T) {
	lines := collectPipeLines(t, "first\nlast")

	if len(lines) != 2 || lines[1] != "last" {
		t.Fatalf("got %q, want final line %q", lines, "last")
	}
}

func TestReadPipeLinesRejectsOversizedLine(t *testing.T) {
	oversized := strings.Repeat("x", constants.MaxPipeLineSize+1) + "\n"
	outputCh := make(chan OutputLine, 1)
	err := readPipeLines(strings.NewReader(oversized), outputCh, passthroughTransform)
	if err == nil {
		t.Fatal("readPipeLines() error = nil, want line size error")
	}
	if !strings.Contains(err.Error(), "line exceeds") {
		t.Fatalf("readPipeLines() error = %v, want line exceeds message", err)
	}
}

type countingReader struct {
	inner io.Reader
	read  int
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.inner.Read(p)
	c.read += n
	return n, err
}

func TestReadPipeLinesStopsReadingOnceLineCapExceeded(t *testing.T) {
	endless := &countingReader{inner: repeatByteReader{}}
	outputCh := make(chan OutputLine, 1)

	err := readPipeLines(endless, outputCh, passthroughTransform)
	if err == nil {
		t.Fatal("readPipeLines() error = nil, want line size error")
	}

	maxExpected := constants.MaxPipeLineSize + 2*constants.PipeReaderBufferSize
	if endless.read > maxExpected {
		t.Fatalf("read %d bytes before erroring, want at most %d (cap must apply before buffering the whole line)", endless.read, maxExpected)
	}
}

type repeatByteReader struct{}

func (repeatByteReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 'x'
	}
	return len(p), nil
}

func realExitError(t *testing.T) *exec.ExitError {
	t.Helper()
	err := exec.Command("false").Run()
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected *exec.ExitError from false, got %T", err)
	}
	return exitErr
}

func errTransform(verbose bool) LineTransformer {
	return func(line string) []OutputLine {
		return []OutputLine{{Text: line, IsErr: true, Verbose: verbose}}
	}
}

func TestRunPipedCommandExitErrorIncludesStderrTail(t *testing.T) {
	mock := &mockCmd{
		stderr:  "--dangerously-skip-permissions must be accepted in an interactive session first.\n",
		waitErr: realExitError(t),
	}

	err := runPipedCommand("claude", mock, nil, passthroughTransform, errTransform(true))
	if err == nil {
		t.Fatal("runPipedCommand() error = nil, want exit error")
	}
	if !strings.Contains(err.Error(), "--dangerously-skip-permissions must be accepted") {
		t.Fatalf("runPipedCommand() error = %v, want stderr detail included", err)
	}

	var detailErr *ExitDetailError
	if !errors.As(err, &detailErr) {
		t.Fatalf("runPipedCommand() error = %T, want *ExitDetailError", err)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("ExitDetailError should unwrap to *exec.ExitError")
	}
}

func TestRunPipedCommandExitErrorPrefersNonVerboseLines(t *testing.T) {
	mock := &mockCmd{
		stderr:  "node warning: deprecated API\nInvalid API key. Please run /login\n",
		waitErr: realExitError(t),
	}

	stderrTransform := func(line string) []OutputLine {
		return []OutputLine{{Text: line, IsErr: true, Verbose: !strings.Contains(line, "Invalid")}}
	}
	err := runPipedCommand("claude", mock, nil, passthroughTransform, stderrTransform)

	var detailErr *ExitDetailError
	if !errors.As(err, &detailErr) {
		t.Fatalf("runPipedCommand() error = %T, want *ExitDetailError", err)
	}
	if len(detailErr.Detail) != 1 || detailErr.Detail[0] != "Invalid API key. Please run /login" {
		t.Fatalf("Detail = %v, want only the non-verbose error line", detailErr.Detail)
	}
}

func TestRunPipedCommandExitErrorCapsTailLength(t *testing.T) {
	mock := &mockCmd{
		stderr:  "e1\ne2\ne3\ne4\ne5\n",
		waitErr: realExitError(t),
	}

	err := runPipedCommand("claude", mock, nil, passthroughTransform, errTransform(false))

	var detailErr *ExitDetailError
	if !errors.As(err, &detailErr) {
		t.Fatalf("runPipedCommand() error = %T, want *ExitDetailError", err)
	}
	want := []string{"e3", "e4", "e5"}
	if len(detailErr.Detail) != len(want) {
		t.Fatalf("Detail = %v, want last %d lines %v", detailErr.Detail, errorTailLimit, want)
	}
	for i, line := range want {
		if detailErr.Detail[i] != line {
			t.Fatalf("Detail = %v, want %v", detailErr.Detail, want)
		}
	}
}

func TestWrapRunnerErrorIncludesDetail(t *testing.T) {
	detailed := &ExitDetailError{
		exitErr: realExitError(t),
		Detail:  []string{"Invalid API key. Please run /login"},
	}

	err := wrapRunnerError("Claude Code", detailed)
	want := "Claude Code exited with code 1: Invalid API key. Please run /login"
	if err.Error() != want {
		t.Fatalf("wrapRunnerError() = %q, want %q", err.Error(), want)
	}
}

func TestWrapRunnerErrorWithoutDetailKeepsBareExitMessage(t *testing.T) {
	err := wrapRunnerError("Claude Code", &ExitDetailError{exitErr: realExitError(t)})
	want := "Claude Code exited with code 1"
	if err.Error() != want {
		t.Fatalf("wrapRunnerError() = %q, want %q", err.Error(), want)
	}
}
