package workflow

import "ralph/internal/shared/runstate"

const StopReasonDuplicateFindings = runstate.StopReasonDuplicateFindings

type ReviewLoopUpdate struct {
	Checkpoint                 string
	ReviewIteration            int
	ReviewFingerprint          string
	ReviewElapsedMs            int64
	StopReason                 string
	LastReviewTranscriptPath   string
	LastReviewChangedFilesHash string
	RecoveryAttempts           int
	ClearRecoveryAttempts      bool
}

type ReviewLoopUpdater interface {
	Snapshot() (iteration int, fingerprint string, elapsedMs int64, changedFilesHash string)
	Apply(u ReviewLoopUpdate) error
}

type checkpointReader interface {
	Checkpoint() string
}
