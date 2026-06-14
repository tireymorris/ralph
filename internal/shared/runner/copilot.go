package runner

import (
	"context"

	"ralph/internal/shared/config"
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
