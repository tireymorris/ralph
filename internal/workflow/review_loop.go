package workflow

import "ralph/internal/shared/runstate"

const StopReasonDuplicateFindings = runstate.StopReasonDuplicateFindings

type ReviewLoopUpdate = runstate.ReviewLoopUpdate

type ReviewLoopUpdater interface {
	Snapshot() (iteration int, fingerprint string, elapsedMs int64, changedFilesHash string)
	Apply(u ReviewLoopUpdate) error
}

type checkpointReader interface {
	Checkpoint() string
}
