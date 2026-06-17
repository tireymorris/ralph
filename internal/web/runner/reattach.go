package runner

import (
	"context"
	"fmt"

	"ralph/internal/shared/runstate"
	"ralph/internal/web/runs"
	"ralph/internal/workflow/events"
)

// Reattach restores an in-memory workflow session for a run interrupted by process restart.
func (c *RunController) Reattach(ctx context.Context) {
	c.mu.Lock()
	if c.reattaching {
		c.mu.Unlock()
		return
	}
	c.reattaching = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.reattaching = false
		c.mu.Unlock()
	}()

	run, ok := c.registry.Get(c.runID)
	if !ok || runs.IsTerminalStatus(run.Status) {
		return
	}

	switch run.Status {
	case runstate.StatusWaitingClarify:
		questions, err := runs.LastClarifyingQuestions(run.WorkDir, c.runID)
		if err != nil {
			c.EmitEvent(events.EventError{Err: fmt.Errorf("restore clarifying questions: %w", err)})
			return
		}
		if len(questions) == 0 {
			c.ForceResume(ctx)
			return
		}
		c.ResumeWaitingClarify(ctx, run.Prompt, questions)
	case runstate.StatusRunning, runstate.StatusImplementing:
		c.ForceResume(ctx)
	}
}

func (c *RunController) ResumeWaitingClarify(ctx context.Context, userPrompt string, questions []string) {
	c.Session.ResumeWaitingClarify(ctx, userPrompt, questions)
}

func (c *RunController) WaitingForClarify() bool {
	return c.Session.WaitingForClarify()
}

func (c *RunController) StartCheckpointResume(ctx context.Context) {
	c.Session.StartCheckpointResume(ctx)
}
