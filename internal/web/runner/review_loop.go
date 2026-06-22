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

func (r *registryReviewLoop) Checkpoint() string {
	run, ok := r.registry.Get(r.runID)
	if !ok {
		return ""
	}
	return run.Checkpoint
}

func (r *registryReviewLoop) Snapshot() (iteration int, fingerprint string, elapsedMs int64, changedFilesHash string) {
	run, ok := r.registry.Get(r.runID)
	if !ok {
		return 0, "", 0, ""
	}
	return run.ReviewIteration, run.ReviewFingerprint, run.ReviewElapsedMs, run.LastReviewChangedFilesHash
}

func (r *registryReviewLoop) Apply(u workflow.ReviewLoopUpdate) error {
	return r.registry.UpdateReviewLoop(r.runID, u)
}

func (r *registryReviewLoop) LastReviewTranscriptPath() string {
	run, ok := r.registry.Get(r.runID)
	if !ok {
		return ""
	}
	return run.LastReviewTranscriptPath
}

func (r *registryReviewLoop) RecoveryAttempts() int {
	run, ok := r.registry.Get(r.runID)
	if !ok {
		return 0
	}
	return run.RecoveryAttempts
}
