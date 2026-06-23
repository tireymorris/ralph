package workflow

import (
	"context"
	"fmt"

	"ralph/internal/prompt"
	"ralph/internal/shared/constants"
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
	changedFiles = gitdiff.ExcludeReviewArtifacts(changedFiles)

	e.emit(EventCleanupStarted{})
	return e.runReviewAndCleanupRounds(ctx, p, changedFiles)
}

func (e *Executor) runReviewAndCleanupRounds(ctx context.Context, p *prd.PRD, changedFiles []string) error {
	e.resetRecoveryAttempts()
	if err := e.runImplementationReviewForCleanup(ctx, p); err != nil {
		return err
	}
	return e.runCleanupRoundsAfterReview(ctx, p, changedFiles)
}

func (e *Executor) runImplementationReviewForCleanup(ctx context.Context, p *prd.PRD) error {
	_, err := e.runImplementationReview(ctx, p)
	return err
}

func (e *Executor) runCleanupRoundsAfterReview(ctx context.Context, p *prd.PRD, changedFiles []string) error {
	if len(changedFiles) == 0 {
		changedFiles, changedFilesErr := gitdiff.ChangedFiles(e.cfg.WorkDir)
		if changedFilesErr != nil {
			logger.Warn("failed to list changed files before cleanup, skipping cleanup", "error", changedFilesErr)
			e.emit(EventOutput{Output: Output{Text: "Skipping cleanup: could not list changed files"}})
			return nil
		}
		changedFiles = gitdiff.ExcludeReviewArtifacts(changedFiles)
	}
	if len(changedFiles) == 0 {
		e.emit(EventOutput{Output: Output{Text: "Skipping cleanup: no changed files"}})
		e.emit(EventCleanupCompleted{})
		return nil
	}

	for round := 1; round <= constants.MaxCleanupRounds; round++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		changedFiles, err := gitdiff.ChangedFiles(e.cfg.WorkDir)
		if err != nil {
			logger.Warn("failed to list changed files during cleanup", "error", err, "round", round)
			break
		}
		changedFiles = gitdiff.ExcludeReviewArtifacts(changedFiles)
		if len(changedFiles) == 0 {
			break
		}

		beforeHash := gitdiff.HashFiles(changedFiles)
		if round > 1 {
			e.emit(EventOutput{Output: Output{Text: fmt.Sprintf("Cleanup round %d of %d", round, constants.MaxCleanupRounds)}})
		}

		cleanupPrompt := prompt.Cleanup(p.Context, e.cfg.PRDFile, changedFiles)
		if runErr := e.runWithForwardedOutput(ctx, cleanupPrompt); runErr != nil {
			e.emit(EventError{Err: fmt.Errorf("cleanup failed: %w", runErr)})
			return runErr
		}

		e.resetRecoveryAttempts()
		if err := e.runTestGateWithRecovery(ctx, p); err != nil {
			return err
		}

		afterChanged, afterErr := gitdiff.ChangedFiles(e.cfg.WorkDir)
		if afterErr != nil {
			logger.Warn("failed to list changed files after cleanup", "error", afterErr, "round", round)
			break
		}
		afterChanged = gitdiff.ExcludeReviewArtifacts(afterChanged)
		if gitdiff.HashFiles(afterChanged) == beforeHash {
			break
		}
	}

	e.emit(EventCleanupCompleted{})
	return nil
}

func (e *Executor) completeRunAfterCleanup(ctx context.Context, p *prd.PRD) error {
	e.resetRecoveryAttempts()
	if err := e.runTestGateWithRecovery(ctx, p); err != nil {
		return err
	}
	e.emit(EventCompleted{})
	return nil
}
