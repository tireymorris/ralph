package runner

import (
	"bufio"
	"context"
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

	var wg sync.WaitGroup
	errCh := make(chan error, constants.PipeReaderCount)
	wg.Add(constants.PipeReaderCount)
	go func() {
		defer wg.Done()
		errCh <- readPipeLines(stdout, outputCh, stdoutTransform)
	}()
	go func() {
		defer wg.Done()
		errCh <- readPipeLines(stderr, outputCh, stderrTransform)
	}()
	wg.Wait()
	close(errCh)
	for readErr := range errCh {
		if readErr != nil {
			return readErr
		}
	}
	return cmd.Wait()
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
		if len(pending) > 0 && outputCh != nil {
			line := strings.TrimSuffix(string(pending), "\n")
			line = strings.TrimSuffix(line, "\r")
			for _, out := range transform(line) {
				outputCh <- out
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
