package workflow

import (
	"context"
	"fmt"

	"ralph/internal/prompt"
	"ralph/internal/shared/gitdiff"
	"ralph/internal/shared/prd"
)

func (e *Executor) RunCleanup(ctx context.Context, p *prd.PRD) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	e.emit(EventCleanupStarted{})

	changedFiles, _ := gitdiff.ChangedFiles(e.cfg.WorkDir)
	cleanupPrompt := prompt.Cleanup(p.Context, e.cfg.PRDFile, changedFiles)

	if runErr := e.runWithForwardedOutput(ctx, cleanupPrompt); runErr != nil {
		e.emit(EventError{Err: fmt.Errorf("cleanup failed: %w", runErr)})
		return runErr
	}

	e.emit(EventCleanupCompleted{})
	return nil
}
