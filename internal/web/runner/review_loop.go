package runner

import (
	"ralph/internal/web/runs"
	"ralph/internal/workflow"
)

type registryReviewLoop struct {
	registry *runs.Registry
	runID    string
}

func newRegistryReviewLoop(registry *runs.Registry, runID string) workflow.ReviewLoopUpdater {
	return &registryReviewLoop{registry: registry, runID: runID}
}

func (r *registryReviewLoop) Snapshot() (iteration int, fingerprint string, elapsedMs int64, changedFilesHash string) {
	run, ok := r.registry.Get(r.runID)
	if !ok {
		return 0, "", 0, ""
	}
	return run.ReviewIteration, run.ReviewFingerprint, run.ReviewElapsedMs, run.LastReviewChangedFilesHash
}

func (r *registryReviewLoop) Apply(u workflow.ReviewLoopUpdate) error {
	return r.registry.UpdateReviewLoop(r.runID, runs.ReviewLoopUpdate{
		Checkpoint:                 u.Checkpoint,
		ReviewIteration:            u.ReviewIteration,
		ReviewFingerprint:          u.ReviewFingerprint,
		ReviewElapsedMs:            u.ReviewElapsedMs,
		StopReason:                 u.StopReason,
		LastReviewTranscriptPath:   u.LastReviewTranscriptPath,
		LastReviewChangedFilesHash: u.LastReviewChangedFilesHash,
		RecoveryAttempts:           u.RecoveryAttempts,
	})
}
