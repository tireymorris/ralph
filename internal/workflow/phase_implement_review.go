package workflow

import (
	"context"
	"fmt"
	"time"

	"ralph/internal/prompt"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/gitdiff"
	"ralph/internal/shared/logger"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runstate"
	"ralph/internal/workflow/review"
)

func (e *Executor) SetReviewLoop(runID string, updater ReviewLoopUpdater) {
	e.runID = runID
	e.reviewLoop = updater
}

func (e *Executor) runImplementationReview(ctx context.Context, p *prd.PRD) (blocked bool, err error) {
	for round := 0; ; round++ {
		if round >= constants.MaxImplementationReviewRounds {
			e.stopImplementationReview(runstate.StopReasonRecoveryExhausted)
			return false, fmt.Errorf("implementation review: exceeded %d review rounds", constants.MaxImplementationReviewRounds)
		}

		blocked, err = e.runImplementationReviewOnce(ctx, p)
		if err == nil && !blocked {
			e.resetRecoveryAttempts()
			return false, nil
		}

		reason := prompt.RecoveryReasonReviewFindings
		errMsg := ""
		var findings []ImplementationFinding
		if err != nil && isDuplicateFindingsError(err) {
			reason = prompt.RecoveryReasonDuplicateFindings
			errMsg = err.Error()
			findings, _ = e.loadPendingFindings()
		} else if blocked {
			findings = e.pendingReviewFindings
			e.resetRecoveryAttempts()
			e.applyReviewLoopBestEffort(e.implReviewLoopUpdate(e.reviewIteration, e.reviewFingerprint, e.reviewElapsedMs, e.reviewChangedFilesHash))
		} else if err != nil {
			return false, err
		}

		recovered, recErr := e.recoverFromReviewFailure(ctx, p, reason, errMsg, findings)
		if recErr != nil {
			return false, recErr
		}
		if !recovered {
			if err != nil && e.recoveryAttemptsSnapshot() >= constants.MaxRecoveryAttempts {
				e.stopImplementationReview(runstate.StopReasonRecoveryExhausted)
				return false, fmt.Errorf("implementation review: %s", runstate.StopReasonRecoveryExhausted)
			}
			continue
		}
	}
}

func (e *Executor) runImplementationReviewOnce(ctx context.Context, p *prd.PRD) (blocked bool, err error) {
	iteration, prevFingerprint, elapsedMs, prevFilesHash := e.reviewLoopSnapshot()
	iteration++

	changed, err := gitdiff.ChangedFiles(e.cfg.WorkDir)
	if err != nil {
		e.emit(EventError{Err: fmt.Errorf("implementation review: %w", err)})
		return false, err
	}
	changed = gitdiff.ExcludeReviewArtifacts(changed)
	filesHash := gitdiff.HashFiles(changed)

	e.emit(EventImplementationReviewStarted{Iteration: iteration})

	if len(changed) > 0 && prevFilesHash != "" && filesHash == prevFilesHash && prevFingerprint != "" {
		e.applyReviewLoopBestEffort(e.implReviewLoopUpdate(iteration, prevFingerprint, elapsedMs, filesHash))
		return false, duplicateFindingsError()
	}

	reviewCtx := ctx
	if e.cfg.RunnerTimeout > 0 {
		var cancel context.CancelFunc
		reviewCtx, cancel = context.WithTimeout(ctx, e.cfg.RunnerTimeout)
		defer cancel()
	}

	start := time.Now()
	result, err := review.ReviewDiffWithChanged(reviewCtx, review.Params{
		WorkDir:   e.cfg.WorkDir,
		RunID:     e.runID,
		Iteration: iteration,
		PRDFile:   e.cfg.PRDFile,
		Context:   p.Context,
		Runner:    e.runner,
	}, changed)
	elapsedMs += time.Since(start).Milliseconds()
	if err != nil {
		if reviewCtx.Err() == context.DeadlineExceeded {
			err = fmt.Errorf("runner invocation timed out after %s: %w", e.cfg.RunnerTimeout, reviewCtx.Err())
		}
		e.emit(EventError{Err: fmt.Errorf("implementation review: %w", err)})
		return false, err
	}

	fingerprint := review.Fingerprint(result.Findings)
	baseUpdate := e.implReviewLoopUpdate(iteration, fingerprint, elapsedMs, filesHash)
	if prevFingerprint != "" && fingerprint != "" && fingerprint == prevFingerprint {
		e.applyReviewLoopBestEffort(baseUpdate)
		return false, duplicateFindingsError()
	}

	baseUpdate.LastReviewTranscriptPath = result.LastReviewTranscriptPath
	e.applyReviewLoopBestEffort(baseUpdate)
	e.lastReviewTranscriptPath = result.LastReviewTranscriptPath

	if len(result.Findings) > 0 {
		e.pendingReviewFindings = result.Findings
		e.emit(EventImplementationReview{Findings: result.Findings})
		return true, nil
	}

	e.pendingReviewFindings = nil
	e.emit(EventImplementationReviewCompleted{Iteration: iteration, Clean: true})
	return false, nil
}

func (e *Executor) reviewLoopSnapshot() (iteration int, fingerprint string, elapsedMs int64, changedFilesHash string) {
	if e.reviewLoop == nil {
		return e.reviewIteration, e.reviewFingerprint, e.reviewElapsedMs, e.reviewChangedFilesHash
	}
	return e.reviewLoop.Snapshot()
}

func (e *Executor) applyReviewLoopBestEffort(u ReviewLoopUpdate) {
	if err := e.applyReviewLoop(u); err != nil {
		logger.Warn("failed to persist review loop state", "error", err, "checkpoint", u.Checkpoint, "run_id", e.runID)
	}
}

func (e *Executor) applyReviewLoop(u ReviewLoopUpdate) error {
	if u.Checkpoint != "" {
		e.reviewIteration = u.ReviewIteration
		e.reviewFingerprint = u.ReviewFingerprint
		e.reviewElapsedMs = u.ReviewElapsedMs
		e.reviewChangedFilesHash = u.LastReviewChangedFilesHash
	}
	if u.LastReviewTranscriptPath != "" {
		e.lastReviewTranscriptPath = u.LastReviewTranscriptPath
	}
	if u.Checkpoint != "" {
		e.recoveryAttempts = u.RecoveryAttempts
	}
	if e.reviewLoop == nil {
		return nil
	}
	return e.reviewLoop.Apply(u)
}

func (e *Executor) stopImplementationReview(reason string) {
	update := e.implReviewLoopUpdate(e.reviewIteration, e.reviewFingerprint, e.reviewElapsedMs, e.reviewChangedFilesHash)
	update.StopReason = reason
	e.applyReviewLoopBestEffort(update)
}

func (e *Executor) implReviewLoopUpdate(iteration int, fingerprint string, elapsedMs int64, filesHash string) ReviewLoopUpdate {
	return ReviewLoopUpdate{
		Checkpoint:                 runstate.CheckpointImplReview,
		ReviewIteration:            iteration,
		ReviewFingerprint:          fingerprint,
		ReviewElapsedMs:            elapsedMs,
		LastReviewChangedFilesHash: filesHash,
		RecoveryAttempts:           e.recoveryAttempts,
	}
}
