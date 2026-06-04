package workflow

import (
	"context"
	"fmt"
	"time"

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
		baseUpdate := e.implReviewLoopUpdate(iteration, prevFingerprint, elapsedMs, filesHash)
		baseUpdate.StopReason = StopReasonDuplicateFindings
		_ = e.applyReviewLoop(baseUpdate)
		stopErr := fmt.Errorf("implementation review: %s", StopReasonDuplicateFindings)
		e.emit(EventError{Err: stopErr})
		return false, stopErr
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
		baseUpdate.StopReason = StopReasonDuplicateFindings
		_ = e.applyReviewLoop(baseUpdate)
		stopErr := fmt.Errorf("implementation review: %s", StopReasonDuplicateFindings)
		e.emit(EventError{Err: stopErr})
		return false, stopErr
	}

	baseUpdate.LastReviewTranscriptPath = result.LastReviewTranscriptPath
	_ = e.applyReviewLoop(baseUpdate)

	if len(result.Findings) > 0 {
		e.emit(EventImplementationReview{Findings: result.Findings})
		return true, nil
	}

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
	e.reviewIteration = u.ReviewIteration
	e.reviewFingerprint = u.ReviewFingerprint
	e.reviewElapsedMs = u.ReviewElapsedMs
	e.reviewChangedFilesHash = u.LastReviewChangedFilesHash
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
	}
}
