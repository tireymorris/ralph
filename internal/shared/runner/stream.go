package runner

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"ralph/internal/shared/constants"
)

type stdinSetter interface {
	setStdin(io.Reader)
}

func setCmdStdin(cmd CmdInterface, stdin io.Reader) {
	if rc, ok := cmd.(*realCmd); ok {
		rc.Cmd.Stdin = io.NopCloser(stdin)
		return
	}
	if sc, ok := cmd.(stdinSetter); ok {
		sc.setStdin(stdin)
	}
}

type CmdInterface interface {
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	Start() error
	Wait() error
}

type realCmd struct {
	*exec.Cmd
}

func (c *realCmd) StdoutPipe() (io.ReadCloser, error) { return c.Cmd.StdoutPipe() }
func (c *realCmd) StderrPipe() (io.ReadCloser, error) { return c.Cmd.StderrPipe() }
func (c *realCmd) Start() error                       { return c.Cmd.Start() }
func (c *realCmd) Wait() error                        { return c.Cmd.Wait() }

func defaultCmdFunc(workDir string) func(ctx context.Context, name string, args ...string) CmdInterface {
	return func(ctx context.Context, name string, args ...string) CmdInterface {
		cmd := exec.CommandContext(ctx, name, args...)
		if workDir != "" {
			cmd.Dir = workDir
		}
		return &realCmd{cmd}
	}
}

func defaultCmdFuncNoStdin(workDir string) func(ctx context.Context, name string, args ...string) CmdInterface {
	return func(ctx context.Context, name string, args ...string) CmdInterface {
		cmd := exec.CommandContext(ctx, name, args...)
		if workDir != "" {
			cmd.Dir = workDir
		}
		cmd.Stdin = nil
		return &realCmd{cmd}
	}
}

type LineTransformer func(line string) []OutputLine

const errorTailLimit = 3

// errorTail keeps the last few error lines a runner emitted so exit errors can
// explain why the process died; CLIs often print the reason only to stderr,
// where the default filter hides it as verbose.
type errorTail struct {
	mu       sync.Mutex
	primary  []string
	fallback []string
}

func (t *errorTail) record(out OutputLine) {
	if !out.IsErr {
		return
	}
	text := strings.TrimSpace(out.Text)
	if text == "" {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	dest := &t.fallback
	if !out.Verbose {
		dest = &t.primary
	}
	*dest = append(*dest, text)
	if len(*dest) > errorTailLimit {
		*dest = (*dest)[1:]
	}
}

func (t *errorTail) snapshot() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	lines := t.primary
	if len(lines) == 0 {
		lines = t.fallback
	}
	return append([]string(nil), lines...)
}

func (t *errorTail) recording(transform LineTransformer) LineTransformer {
	return func(line string) []OutputLine {
		outs := transform(line)
		for _, out := range outs {
			t.record(out)
		}
		return outs
	}
}

// ExitDetailError pairs a process exit error with the error lines that explain it.
type ExitDetailError struct {
	exitErr *exec.ExitError
	Detail  []string
}

func (e *ExitDetailError) Error() string {
	if len(e.Detail) == 0 {
		return e.exitErr.Error()
	}
	return fmt.Sprintf("%s: %s", e.exitErr.Error(), strings.Join(e.Detail, " | "))
}

func (e *ExitDetailError) Unwrap() error { return e.exitErr }

func (e *ExitDetailError) ExitCode() int { return e.exitErr.ExitCode() }

// runPipedCommand streams stdout/stderr through transformers before waiting on cmd.
func runPipedCommand(commandName string, cmd CmdInterface, outputCh chan<- OutputLine, stdoutTransform, stderrTransform LineTransformer) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe for %s: %w", commandName, err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe for %s: %w", commandName, err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", commandName, err)
	}

	tail := &errorTail{}
	var wg sync.WaitGroup
	errCh := make(chan error, constants.PipeReaderCount)
	wg.Add(constants.PipeReaderCount)
	go func() {
		defer wg.Done()
		errCh <- readPipeLines(stdout, outputCh, tail.recording(stdoutTransform))
	}()
	go func() {
		defer wg.Done()
		errCh <- readPipeLines(stderr, outputCh, tail.recording(stderrTransform))
	}()
	wg.Wait()
	close(errCh)
	for readErr := range errCh {
		if readErr != nil {
			return readErr
		}
	}
	if waitErr := cmd.Wait(); waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			return &ExitDetailError{exitErr: exitErr, Detail: tail.snapshot()}
		}
		return waitErr
	}
	return nil
}

// readPipeLines reads lines longer than any fixed bufio.Scanner buffer (AI
// runners emit NDJSON events embedding diffs), accumulating buffer-sized
// fragments so the MaxPipeLineSize cap is enforced before a pathological
// line is fully buffered.
func readPipeLines(pipe io.Reader, outputCh chan<- OutputLine, transform LineTransformer) error {
	reader := bufio.NewReaderSize(pipe, constants.PipeReaderBufferSize)
	var pending []byte
	for {
		chunk, err := reader.ReadSlice('\n')
		pending = append(pending, chunk...)
		if len(pending) > constants.MaxPipeLineSize {
			return fmt.Errorf("scan pipe output: line exceeds %d bytes", constants.MaxPipeLineSize)
		}
		if err == bufio.ErrBufferFull {
			continue
		}
		if len(pending) > 0 {
			line := strings.TrimSuffix(string(pending), "\n")
			line = strings.TrimSuffix(line, "\r")
			for _, out := range transform(line) {
				if outputCh != nil {
					outputCh <- out
				}
			}
		}
		pending = pending[:0]
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("scan pipe output: %w", err)
		}
	}
}
