package workflow

import (
	"context"
	"fmt"

	"ralph/internal/prompt"
	"ralph/internal/shared/gitdiff"
	"ralph/internal/shared/logger"
	"ralph/internal/shared/prd"
)

func (e *Executor) RunCleanup(ctx context.Context, p *prd.PRD) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	changedFiles, changedFilesErr := gitdiff.ChangedFiles(e.cfg.WorkDir)
	if changedFilesErr != nil {
		logger.Warn("failed to list changed files before cleanup, skipping cleanup", "error", changedFilesErr)
		e.emit(EventOutput{Output: Output{Text: "Skipping cleanup: could not list changed files"}})
		return nil
	}
	if len(changedFiles) == 0 {
		e.emit(EventOutput{Output: Output{Text: "Skipping cleanup: no changed files"}})
		return nil
	}

	e.emit(EventCleanupStarted{})

	cleanupPrompt := prompt.Cleanup(p.Context, e.cfg.PRDFile, changedFiles)

	if runErr := e.runWithForwardedOutput(ctx, cleanupPrompt); runErr != nil {
		e.emit(EventError{Err: fmt.Errorf("cleanup failed: %w", runErr)})
		return runErr
	}

	e.emit(EventCleanupCompleted{})
	return nil
}
