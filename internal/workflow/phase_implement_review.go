package workflow

import (
	"context"
	"fmt"
	"time"

	"ralph/internal/prompt"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/gitdiff"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runstate"
	"ralph/internal/workflow/review"
)

func (e *Executor) SetReviewLoop(runID string, updater ReviewLoopUpdater) {
	e.runID = runID
	e.reviewLoop = updater
}

func (e *Executor) runImplementationReview(ctx context.Context, p *prd.PRD) (blocked bool, err error) {
	for {
		blocked, err = e.runImplementationReviewOnce(ctx, p)
		if err == nil && !blocked {
			e.recoveryAttempts = 0
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
		} else if err != nil {
			return false, err
		}

		recovered, recErr := e.recoverFromReviewFailure(ctx, p, reason, errMsg, findings)
		if recErr != nil {
			return blocked, recErr
		}
		if !recovered {
			if e.recoveryAttemptsSnapshot() >= constants.MaxRecoveryAttempts {
				if err != nil {
					baseUpdate := e.implReviewLoopUpdate(e.reviewIteration, e.reviewFingerprint, e.reviewElapsedMs, e.reviewChangedFilesHash)
					baseUpdate.StopReason = runstate.StopReasonRecoveryExhausted
					_ = e.applyReviewLoop(baseUpdate)
					return false, fmt.Errorf("implementation review: %s", runstate.StopReasonRecoveryExhausted)
				}
				return blocked, nil
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
	filesHash := gitdiff.HashFiles(changed)

	e.emit(EventImplementationReviewStarted{Iteration: iteration})

	if len(changed) > 0 && prevFilesHash != "" && filesHash == prevFilesHash && prevFingerprint != "" {
		_ = e.applyReviewLoop(e.implReviewLoopUpdate(iteration, prevFingerprint, elapsedMs, filesHash))
		return false, duplicateFindingsError()
	}

	start := time.Now()
	result, err := review.ReviewDiffWithChanged(ctx, review.Params{
		WorkDir:   e.cfg.WorkDir,
		RunID:     e.runID,
		Iteration: iteration,
		PRDFile:   e.cfg.PRDFile,
		Context:   p.Context,
		Runner:    e.runner,
	}, changed)
	elapsedMs += time.Since(start).Milliseconds()
	if err != nil {
		e.emit(EventError{Err: fmt.Errorf("implementation review: %w", err)})
		return false, err
	}

	fingerprint := review.Fingerprint(result.Findings)
	baseUpdate := e.implReviewLoopUpdate(iteration, fingerprint, elapsedMs, filesHash)
	if prevFingerprint != "" && fingerprint != "" && fingerprint == prevFingerprint {
		_ = e.applyReviewLoop(baseUpdate)
		return false, duplicateFindingsError()
	}

	baseUpdate.LastReviewTranscriptPath = result.LastReviewTranscriptPath
	_ = e.applyReviewLoop(baseUpdate)
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
	if u.RecoveryAttempts > 0 {
		e.recoveryAttempts = u.RecoveryAttempts
	}
	if e.reviewLoop == nil {
		return nil
	}
	return e.reviewLoop.Apply(u)
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
