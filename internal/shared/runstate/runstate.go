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
	StatusWaitingImplReview = "waiting_implementation_review"
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

const CheckpointCancelled = "cancelled"

func EventStatusPhase(ev events.Event) (status, phase string) {
	switch ev.(type) {
	case events.EventClarifyingQuestions:
		return "waiting_clarify", PhaseClarify
	case events.EventPRDGenerating, events.EventPRDRevising:
		return "running", PhaseGenerate
	case events.EventPRDGenerated, events.EventPRDLoaded, events.EventPRDReview:
		return "waiting_review", PhaseReview
	case events.EventStoryStarted, events.EventStoryCompleted:
		return "implementing", PhaseImplement
	case events.EventImplementationReview:
		return StatusWaitingImplReview, PhaseImplementationReview
	case events.EventCleanupStarted, events.EventCleanupCompleted:
		return "implementing", PhaseCleanup
	case events.EventCompleted:
		return "completed", PhaseCompleted
	case events.EventError:
		return "failed", PhaseFailed
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
		return "running", PhaseGenerate
	}
	return "implementing", PhaseImplement
}
