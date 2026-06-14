package runner

import (
	"context"
	"strings"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/logger"
)

type CopilotRunner struct {
	cfg     *config.Config
	CmdFunc func(ctx context.Context, name string, args ...string) CmdInterface
}

func NewCopilot(cfg *config.Config) *CopilotRunner {
	return &CopilotRunner{
		cfg:     cfg,
		CmdFunc: defaultCmdFuncNoStdin(cfg.WorkDir),
	}
}

func (r *CopilotRunner) RunnerName() string {
	return "copilot"
}

func (r *CopilotRunner) CommandName() string {
	return "copilot"
}

func (r *CopilotRunner) IsInternalLog(line string) bool {
	return stderrLineIsInternal(line, stderrFilterDefaultPipedCLI)
}

func (r *CopilotRunner) Run(ctx context.Context, prompt string, outputCh chan<- OutputLine) error {
	args := []string{"--allow-all-tools", "--allow-all-paths", "--no-ask-user", "--output-format", "json", "--autopilot"}

	logger.Debug("invoking AI runner",
		"runner", r.RunnerName(),
		"command", r.CommandName(),
		"runner", r.cfg.Runner,
		"prompt_length", len(prompt),
		"work_dir", r.cfg.WorkDir)

	if outputCh != nil {
		outputCh <- newStartingOutputLine(r.RunnerName())
	}

	err := runWithPipedCommandAndStdin(ctx, r.CommandName(), r.CmdFunc, strings.NewReader(prompt), args, outputCh,
		parseCopilotJSONL,
		func(line string) []OutputLine {
			return []OutputLine{{Text: line, IsErr: true, Time: time.Now(), Verbose: r.IsInternalLog(line)}}
		},
	)

	if err != nil {
		logger.Debug("AI runner exited with code",
			"runner", r.RunnerName(),
			"command", r.CommandName(),
			"runner", r.cfg.Runner)
		return wrapRunnerError(r.RunnerName(), err)
	}

	logger.Debug("AI runner completed successfully",
		"runner", r.RunnerName(),
		"command", r.CommandName(),
		"runner", r.cfg.Runner)
	return nil
}

func parseCopilotJSONL(line string) []OutputLine {
	return nil
}
