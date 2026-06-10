package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ralph/internal/prompt"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/gitdiff"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runstate"
	"ralph/internal/workflow/events"
	"ralph/internal/workflow/review"
)

func isDuplicateFindingsError(err error) bool {
	return err != nil && strings.Contains(err.Error(), runstate.StopReasonDuplicateFindings)
}

func duplicateFindingsError() error {
	return fmt.Errorf("implementation review: %s", runstate.StopReasonDuplicateFindings)
}

func recoveryFindingsFromEvents(findings []ImplementationFinding) []prompt.RecoveryFinding {
	out := make([]prompt.RecoveryFinding, 0, len(findings))
	for _, f := range findings {
		out = append(out, prompt.RecoveryFinding{
			Category: f.Category,
			Path:     f.Path,
			Line:     f.Line,
			Summary:  f.Summary,
		})
	}
	return out
}

func (e *Executor) recoveryAttemptsSnapshot() int {
	if e.reviewLoop == nil {
		return e.recoveryAttempts
	}
	if ext, ok := e.reviewLoop.(recoveryAttemptsReader); ok {
		return ext.RecoveryAttempts()
	}
	return e.recoveryAttempts
}

func (e *Executor) applyMechanicalCleanup(findings []ImplementationFinding) {
	for _, f := range findings {
		if f.Path == "" {
			continue
		}
		switch f.Category {
		case "wrong-target", "process", "artifact":
			removed, err := gitdiff.RemoveUntracked(e.cfg.WorkDir, f.Path)
			if err == nil && removed {
				e.emit(EventOutput{Output: events.Output{Text: fmt.Sprintf("Removed untracked artifact: %s", f.Path)}})
			}
		}
	}
}

func (e *Executor) clearReviewFingerprint() {
	iteration, _, elapsed, filesHash := e.reviewLoopSnapshot()
	e.applyReviewLoopBestEffort(ReviewLoopUpdate{
		Checkpoint:                 runstate.CheckpointImplReview,
		ReviewIteration:            iteration,
		ReviewFingerprint:          "",
		ReviewElapsedMs:            elapsed,
		LastReviewChangedFilesHash: filesHash,
		RecoveryAttempts:           e.recoveryAttempts,
	})
}

func (e *Executor) loadPendingFindings() ([]ImplementationFinding, error) {
	transcriptPath := e.lastReviewTranscriptPath
	if transcriptPath == "" && e.reviewLoop != nil {
		if reader, ok := e.reviewLoop.(transcriptPathReader); ok {
			transcriptPath = reader.LastReviewTranscriptPath()
		}
	}
	if transcriptPath == "" {
		return nil, nil
	}
	path := filepath.Join(e.cfg.WorkDir, ".ralph", "runs", e.runID, transcriptPath)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	findings, err := review.ParseFindings(string(data), false)
	if err != nil {
		return nil, err
	}
	return findings, nil
}

func (e *Executor) runRecovery(
	ctx context.Context,
	p *prd.PRD,
	reason prompt.RecoveryReason,
	errMsg string,
	findings []ImplementationFinding,
) (bool, error) {
	attempts := e.recoveryAttemptsSnapshot()
	if attempts >= constants.MaxRecoveryAttempts {
		return false, nil
	}

	attempt := attempts + 1
	e.emit(EventRecoveryStarted{
		Reason:  string(reason),
		Attempt: attempt,
		Max:     constants.MaxRecoveryAttempts,
	})

	e.applyMechanicalCleanup(findings)

	changed, err := gitdiff.ChangedFiles(e.cfg.WorkDir)
	if err != nil {
		return false, err
	}

	recoveryPrompt := prompt.RecoverFromFailure(
		p.Context,
		e.cfg.PRDFile,
		reason,
		attempt,
		constants.MaxRecoveryAttempts,
		errMsg,
		recoveryFindingsFromEvents(findings),
		changed,
	)

	runErr := e.runWithForwardedOutput(ctx, recoveryPrompt)
	success := runErr == nil
	e.emit(EventRecoveryCompleted{
		Reason:  string(reason),
		Attempt: attempt,
		Success: success,
	})

	e.recoveryAttempts = attempt
	u := ReviewLoopUpdate{RecoveryAttempts: attempt}
	if e.reviewLoop != nil {
		iteration, fingerprint, elapsed, filesHash := e.reviewLoop.Snapshot()
		u.Checkpoint = runstate.CheckpointImplReview
		u.ReviewIteration = iteration
		u.ReviewFingerprint = fingerprint
		u.ReviewElapsedMs = elapsed
		u.LastReviewChangedFilesHash = filesHash
	}
	e.applyReviewLoopBestEffort(u)

	if !success {
		return false, runErr
	}
	return true, nil
}

func (e *Executor) recoverFromReviewFailure(
	ctx context.Context,
	p *prd.PRD,
	reason prompt.RecoveryReason,
	errMsg string,
	findings []ImplementationFinding,
) (bool, error) {
	beforeChanged, err := gitdiff.ChangedFiles(e.cfg.WorkDir)
	if err != nil {
		return false, err
	}
	beforeHash := gitdiff.HashFiles(beforeChanged)

	recovered, err := e.runRecovery(ctx, p, reason, errMsg, findings)
	if err != nil {
		return false, err
	}
	if !recovered {
		return false, nil
	}

	afterChanged, chErr := gitdiff.ChangedFiles(e.cfg.WorkDir)
	if chErr != nil {
		return false, chErr
	}
	afterHash := gitdiff.HashFiles(afterChanged)
	agentProgress := afterHash != beforeHash

	if agentProgress {
		committed, commitErr := gitdiff.CommitTrackedChanges(e.cfg.WorkDir, "ralph: recovery fixes")
		if commitErr != nil {
			return false, commitErr
		}
		if committed {
			e.emit(EventOutput{Output: events.Output{Text: "Committed recovery fixes before re-review."}})
		}
		e.clearReviewFingerprint()
		return true, nil
	}

	if reason == prompt.RecoveryReasonDuplicateFindings {
		return false, nil
	}

	committed, commitErr := gitdiff.CommitTrackedChanges(e.cfg.WorkDir, "ralph: recovery fixes")
	if commitErr != nil {
		return false, commitErr
	}
	if !committed {
		return false, nil
	}

	e.emit(EventOutput{Output: events.Output{Text: "Committed recovery fixes before re-review."}})
	e.clearReviewFingerprint()
	return true, nil
}

func (e *Executor) RunImplementationAfterReviewRecovery(ctx context.Context, p *prd.PRD) error {
	findings, err := e.loadPendingFindings()
	if err != nil {
		e.emit(EventError{Err: fmt.Errorf("load review findings: %w", err)})
		return err
	}
	recovered, recErr := e.recoverFromReviewFailure(ctx, p, prompt.RecoveryReasonManualContinue, "", findings)
	if recErr != nil {
		e.emit(EventError{Err: recErr})
		return recErr
	}
	if recovered {
		blocked, reviewErr := e.runImplementationReview(ctx, p)
		if reviewErr != nil {
			return reviewErr
		}
		if blocked {
			return nil
		}
	}
	return e.RunImplementation(ctx, p)
}

type recoveryAttemptsReader interface {
	RecoveryAttempts() int
}

type transcriptPathReader interface {
	LastReviewTranscriptPath() string
}
