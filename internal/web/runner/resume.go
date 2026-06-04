package runner

import (
	"context"
	"fmt"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/web/runs"
	"ralph/internal/workflow/events"
)

func (c *RunController) runConfig() *config.Config {
	runCfg := *c.cfg
	if run, ok := c.registry.Get(c.runID); ok {
		runCfg.WorkDir = run.WorkDir
		if run.PRDPath != "" {
			runCfg.PRDFile = run.PRDPath
		}
	}
	return &runCfg
}

// ForceResume cancels the current stuck step and continues from on-disk state.
func (c *RunController) ForceResume(ctx context.Context) {
	run, ok := c.registry.Get(c.runID)
	if !ok {
		c.EmitEvent(events.EventError{Err: fmt.Errorf("run %s not found", c.runID)})
		return
	}

	runCfg := c.runConfig()
	p, err := prd.Load(runCfg)
	if err == nil {
		switch run.Checkpoint {
		case runs.CheckpointPRDReview:
			c.Driver.StartResume(ctx)
			return
		case runs.CheckpointImplReview, runs.CheckpointFollowup:
			c.Driver.StartImplementation(ctx, p)
			return
		case runs.CheckpointComplete:
			return
		}
		if !p.AllCompleted() && run.Status != "waiting_review" {
			c.Driver.StartImplementation(ctx, p)
			return
		}
		c.Driver.StartResume(ctx)
		return
	}
	if run.Prompt != "" {
		c.StartNew(ctx, run.Prompt)
		return
	}
	c.EmitEvent(events.EventError{Err: fmt.Errorf("cannot resume: no PRD or prompt")})
}
