package runstate

import (
	"ralph/internal/shared/prd"
	"ralph/internal/workflow/events"
)

const LocalRunID = "prd-local"

const (
	CheckpointPRDReview  = "prd_review"
	CheckpointImplReview = "impl_review"
	CheckpointFollowup   = "followup"
	CheckpointComplete   = "complete"
)

const (
	StatusRunning           = "running"
	StatusWaitingClarify    = "waiting_clarify"
	StatusWaitingReview     = "waiting_review"
	StatusWaitingImplReview = "waiting_implementation_review"
	StatusImplementing      = "implementing"
	StatusCompleted         = "completed"
	StatusFailed            = "failed"
	StatusCancelled         = "cancelled"
)

const StopReasonDuplicateFindings = "duplicate_findings"

const (
	PhaseGenerate             = "generate"
	PhaseClarify              = "clarify"
	PhaseReview               = "review"
	PhaseImplement            = "implement"
	PhaseImplementationReview = "implementation_review"
	PhaseFollowup             = "followup"
	PhaseCleanup              = "cleanup"
	PhaseCompleted            = "completed"
	PhaseFailed               = "failed"
	PhaseCancelled            = "cancelled"
)

const CheckpointCancelled = StatusCancelled

func EventStatusPhase(ev events.Event) (status, phase string) {
	switch ev.(type) {
	case events.EventClarifyingQuestions:
		return StatusWaitingClarify, PhaseClarify
	case events.EventPRDGenerating, events.EventPRDRevising:
		return StatusRunning, PhaseGenerate
	case events.EventPRDGenerated, events.EventPRDLoaded, events.EventPRDReview:
		return StatusWaitingReview, PhaseReview
	case events.EventStoryStarted, events.EventStoryCompleted:
		return StatusImplementing, PhaseImplement
	case events.EventImplementationReview:
		return StatusWaitingImplReview, PhaseImplementationReview
	case events.EventCleanupStarted, events.EventCleanupCompleted:
		return StatusImplementing, PhaseCleanup
	case events.EventCompleted:
		return StatusCompleted, PhaseCompleted
	case events.EventError:
		return StatusFailed, PhaseFailed
	default:
		return "", ""
	}
}

func EventCheckpoint(ev events.Event) string {
	switch ev.(type) {
	case events.EventPRDReview:
		return CheckpointPRDReview
	case events.EventImplementationReviewStarted, events.EventImplementationReview, events.EventImplementationReviewCompleted:
		return CheckpointImplReview
	case events.EventCompleted:
		return CheckpointComplete
	default:
		return ""
	}
}

func CheckpointPhase(checkpoint string, p *prd.PRD) string {
	switch checkpoint {
	case CheckpointPRDReview:
		return PhaseReview
	case CheckpointImplReview:
		return PhaseImplementationReview
	case CheckpointFollowup:
		return PhaseFollowup
	case CheckpointComplete:
		return PhaseCompleted
	case CheckpointCancelled:
		return PhaseCancelled
	default:
		if p != nil && !p.AllCompleted() {
			return PhaseImplement
		}
		return PhaseReview
	}
}

func LocalPRDStatusPhase(p *prd.PRD, checkpoint string) (status, phase string) {
	if checkpoint == CheckpointImplReview {
		return StatusWaitingImplReview, PhaseImplementationReview
	}
	if p == nil || len(p.Stories) == 0 {
		return StatusRunning, PhaseGenerate
	}
	return StatusImplementing, PhaseImplement
}
