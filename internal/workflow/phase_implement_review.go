package workflow

import (
	"context"
	"fmt"
	"time"

	"ralph/internal/shared/prd"
	"ralph/internal/workflow/review"
)

func (e *Executor) SetReviewLoop(runID string, updater ReviewLoopUpdater) {
	e.runID = runID
	e.reviewLoop = updater
}

func (e *Executor) runImplementationReview(ctx context.Context, p *prd.PRD) (blocked bool, err error) {
	iteration, prevFingerprint, elapsedMs := e.reviewLoopSnapshot()
	iteration++

	e.emit(EventImplementationReviewStarted{Iteration: iteration})

	start := time.Now()
	result, err := review.ReviewDiff(ctx, review.Params{
		WorkDir:   e.cfg.WorkDir,
		RunID:     e.runID,
		Iteration: iteration,
		PRDFile:   e.cfg.PRDFile,
		Context:   p.Context,
		Runner:    e.runner,
	})
	elapsedMs += time.Since(start).Milliseconds()
	if err != nil {
		e.emit(EventError{Err: fmt.Errorf("implementation review: %w", err)})
		return false, err
	}

	fingerprint := review.Fingerprint(result.Findings)
	baseUpdate := e.implReviewLoopUpdate(iteration, fingerprint, elapsedMs)
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

func (e *Executor) reviewLoopSnapshot() (iteration int, fingerprint string, elapsedMs int64) {
	if e.reviewLoop == nil {
		return e.reviewIteration, e.reviewFingerprint, e.reviewElapsedMs
	}
	return e.reviewLoop.Snapshot()
}

func (e *Executor) applyReviewLoop(u ReviewLoopUpdate) error {
	e.reviewIteration = u.ReviewIteration
	e.reviewFingerprint = u.ReviewFingerprint
	e.reviewElapsedMs = u.ReviewElapsedMs
	if e.reviewLoop == nil {
		return nil
	}
	return e.reviewLoop.Apply(u)
}

func (e *Executor) implReviewLoopUpdate(iteration int, fingerprint string, elapsedMs int64) ReviewLoopUpdate {
	return ReviewLoopUpdate{
		Checkpoint:        CheckpointImplReview,
		ReviewIteration:   iteration,
		ReviewFingerprint: fingerprint,
		ReviewElapsedMs:   elapsedMs,
	}
}
