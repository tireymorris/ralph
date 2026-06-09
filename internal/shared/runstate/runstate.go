package runstate

import "ralph/internal/shared/prd"

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

const (
	StopReasonDuplicateFindings = "duplicate_findings"
	StopReasonRecoveryExhausted = "recovery_exhausted"
)

const (
	PhaseGenerate             = "generate"
	PhaseClarify              = "clarify"
	PhaseReview               = "review"
	PhaseImplement            = "implement"
	PhaseImplementationReview = "implementation_review"
	PhaseFollowup             = "followup"
	PhaseCleanup              = "cleanup"
	PhaseCompleted            = "complete"
	PhaseFailed               = "failed"
	PhaseCancelled            = "cancelled"
)

const CheckpointCancelled = StatusCancelled

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
		return StatusWaitingImplReview, PhaseImplement
	}
	if p == nil || len(p.Stories) == 0 {
		return StatusRunning, PhaseGenerate
	}
	return StatusImplementing, PhaseImplement
}
