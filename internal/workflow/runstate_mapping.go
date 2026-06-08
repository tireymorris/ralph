package workflow

import (
	"ralph/internal/shared/runstate"
	"ralph/internal/workflow/events"
)

func EventStatusPhase(ev events.Event) (status, phase string) {
	switch ev.(type) {
	case events.EventClarifyingQuestions:
		return runstate.StatusWaitingClarify, runstate.PhaseClarify
	case events.EventPRDGenerating, events.EventPRDRevising:
		return runstate.StatusRunning, runstate.PhaseGenerate
	case events.EventPRDGenerated, events.EventPRDLoaded, events.EventPRDReview:
		return runstate.StatusWaitingReview, runstate.PhaseReview
	case events.EventStoryStarted, events.EventStoryCompleted:
		return runstate.StatusImplementing, runstate.PhaseImplement
	case events.EventImplementationReview:
		return runstate.StatusWaitingImplReview, runstate.PhaseImplement
	case events.EventCleanupStarted, events.EventCleanupCompleted:
		return runstate.StatusImplementing, runstate.PhaseCleanup
	case events.EventCompleted:
		return runstate.StatusCompleted, runstate.PhaseCompleted
	case events.EventError:
		return runstate.StatusFailed, runstate.PhaseFailed
	default:
		return "", ""
	}
}

func EventCheckpoint(ev events.Event) string {
	switch ev.(type) {
	case events.EventPRDReview:
		return runstate.CheckpointPRDReview
	case events.EventImplementationReviewStarted, events.EventImplementationReview, events.EventImplementationReviewCompleted:
		return runstate.CheckpointImplReview
	case events.EventCompleted:
		return runstate.CheckpointComplete
	default:
		return ""
	}
}
