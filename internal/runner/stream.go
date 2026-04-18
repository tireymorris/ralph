package runner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"ralph/internal/constants"
)

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

// runPipedCommand starts cmd, streams stdout/stderr through transform, waits for readers, then Wait.
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
	wg.Add(constants.PipeReaderCount)
	go func() {
		defer wg.Done()
		readPipeLines(stdout, outputCh, stdoutTransform)
	}()
	go func() {
		defer wg.Done()
		readPipeLines(stderr, outputCh, stderrTransform)
	}()
	wg.Wait()
	return cmd.Wait()
}

func readPipeLines(pipe io.Reader, outputCh chan<- OutputLine, transform LineTransformer) {
	scanner := bufio.NewScanner(pipe)
	buf := make([]byte, 0, constants.InitialScannerBufferCapacity)
	scanner.Buffer(buf, constants.ScannerBufferSize)
	for scanner.Scan() {
		if outputCh != nil {
			for _, out := range transform(scanner.Text()) {
				outputCh <- out
			}
		}
	}
}
