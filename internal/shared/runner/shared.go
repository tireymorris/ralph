package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

func newStartingOutputLine(runnerName string) OutputLine {
	return OutputLine{Text: fmt.Sprintf("Starting %s...", runnerName), Time: time.Now()}
}

func runWithPipedCommandAndStdin(
	ctx context.Context,
	cmdName string,
	cmdFactory func(context.Context, string, ...string) CmdInterface,
	stdin io.Reader,
	args []string,
	outputCh chan<- OutputLine,
	stdoutTransform, stderrTransform LineTransformer,
) error {
	cmd := cmdFactory(ctx, cmdName, args...)
	setCmdStdin(cmd, stdin)
	return runPipedCommand(cmdName, cmd, outputCh, stdoutTransform, stderrTransform)
}

func wrapRunnerError(runnerName string, err error) error {
	var detailErr *ExitDetailError
	if errors.As(err, &detailErr) {
		if len(detailErr.Detail) > 0 {
			return fmt.Errorf("%s exited with code %d: %s", runnerName, detailErr.ExitCode(), strings.Join(detailErr.Detail, " | "))
		}
		return fmt.Errorf("%s exited with code %d", runnerName, detailErr.ExitCode())
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return fmt.Errorf("%s exited with code %d", runnerName, exitErr.ExitCode())
	}
	return fmt.Errorf("%s failed: %w", runnerName, err)
}

func exitCode(err error) int {
	var detailErr *ExitDetailError
	if errors.As(err, &detailErr) {
		return detailErr.ExitCode()
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}
