package runner

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

func newStartingOutputLine(runnerName string) OutputLine {
	return OutputLine{Text: fmt.Sprintf("Starting %s...", runnerName), Time: time.Now()}
}

func runWithPipedCommand(
	ctx context.Context,
	cmdName string,
	cmdFactory func(context.Context, string, ...string) CmdInterface,
	args []string,
	outputCh chan<- OutputLine,
	stdoutTransform, stderrTransform LineTransformer,
) error {
	cmd := cmdFactory(ctx, cmdName, args...)
	return runPipedCommand(cmdName, cmd, outputCh, stdoutTransform, stderrTransform)
}

func wrapRunnerError(runnerName string, err error) error {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("%s exited with code %d", runnerName, exitErr.ExitCode())
	}
	return fmt.Errorf("%s failed: %w", runnerName, err)
}
