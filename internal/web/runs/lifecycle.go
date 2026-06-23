package runs

import "ralph/internal/shared/runstate"

type Lifecycle struct {
	registry *Registry
}

func NewLifecycle(registry *Registry) *Lifecycle {
	return &Lifecycle{registry: registry}
}

func (l *Lifecycle) Clarify(id string) error {
	return l.registry.UpdateLifecycle(id, LifecycleUpdate{
		Status: runstate.StatusRunning,
		Phase:  runstate.PhaseGenerate,
	})
}

func (l *Lifecycle) ReviseReview(id string) error {
	return l.registry.UpdateLifecycle(id, LifecycleUpdate{
		Status: runstate.StatusRunning,
		Phase:  runstate.PhaseGenerate,
	})
}

func (l *Lifecycle) ContinueImplementationReview(id string) error {
	return l.registry.UpdateLifecycle(id, LifecycleUpdate{
		Status: runstate.StatusImplementing,
		Phase:  runstate.PhaseCleanup,
	})
}

func (l *Lifecycle) FollowUp(id string) error {
	return l.registry.UpdateLifecycle(id, LifecycleUpdate{
		Status:     runstate.StatusRunning,
		Phase:      runstate.PhaseFollowup,
		Checkpoint: runstate.CheckpointFollowup,
	})
}
