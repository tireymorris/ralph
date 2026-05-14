// Package review implements the "ralph review" subcommand.
package review

import (
	"context"
	"fmt"
	"os"

	"ralph/internal/shared/config"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/logger"
	"ralph/internal/shared/prd"
	"ralph/internal/prompt"
	"ralph/internal/workflow"
)

type workflowExecutor interface {
	RunClarify(ctx context.Context, userPrompt string) ([]prompt.QuestionAnswer, error)
	RunLoad(ctx context.Context) (*prd.PRD, error)
	RunGenerateWithAnswers(ctx context.Context, userPrompt string, qas []prompt.QuestionAnswer) (*prd.PRD, error)
}

// Cmd runs the review subcommand workflow.
type Cmd struct {
	cfg      *config.Config
	verbose  bool
	executor workflowExecutor
	eventsCh chan workflow.Event
}

// NewCmd creates a new Cmd instance.
func NewCmd(cfg *config.Config, verbose bool) *Cmd {
	eventsCh := make(chan workflow.Event, constants.EventChannelBuffer)
	return &Cmd{
		cfg:      cfg,
		verbose:  verbose,
		eventsCh: eventsCh,
		executor: workflow.NewExecutor(cfg, eventsCh),
	}
}

// Run loads existing prd.json, runs clarification if needed, presents review prompt, and exits.
func (c *Cmd) Run() int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exists, err := prd.Exists(c.cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking PRD file: %v\n", err)
		return 1
	}
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: no PRD file found. Run 'ralph prd <prompt>' first.\n")
		return 1
	}

	p, err := c.executor.RunLoad(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading PRD: %v\n", err)
		return 1
	}

	logger.Debug("PRD loaded for review", "project", p.ProjectName, "stories", len(p.Stories))

	close(c.eventsCh)
	return 0
}

// Run is a convenience function that creates a Cmd and runs it.
func Run(cfg *config.Config, verbose bool) int {
	return NewCmd(cfg, verbose).Run()
}
