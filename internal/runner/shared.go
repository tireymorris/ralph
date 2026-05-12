package runner

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

func newStartingOutputLine(runnerName, model string) OutputLine {
	return OutputLine{Text: fmt.Sprintf("Starting %s with model %s...", runnerName, model), Time: time.Now()}
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

func wrapRunnerError(runnerName, model string, err error) error {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("%s with model %s exited with code %d", runnerName, model, exitErr.ExitCode())
	}
	return fmt.Errorf("%s with model %s failed: %w", runnerName, model, err)
}
