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

func CheckpointStatusPhase(checkpoint string, p *prd.PRD) (status, phase string) {
	switch checkpoint {
	case CheckpointPRDReview:
		return StatusWaitingReview, PhaseReview
	case CheckpointImplReview:
		return StatusWaitingImplReview, PhaseImplementationReview
	case CheckpointFollowup:
		return StatusImplementing, PhaseImplement
	case CheckpointComplete:
		return StatusCompleted, PhaseCompleted
	case CheckpointCancelled:
		return StatusCancelled, PhaseCancelled
	default:
		if p != nil && !p.AllCompleted() {
			return StatusImplementing, PhaseImplement
		}
		return StatusRunning, PhaseReview
	}
}

func LocalPRDStatusPhase(p *prd.PRD, checkpoint string) (status, phase string) {
	if checkpoint != "" {
		return CheckpointStatusPhase(checkpoint, p)
	}
	if p == nil || len(p.Stories) == 0 {
		return StatusRunning, PhaseGenerate
	}
	return StatusImplementing, PhaseImplement
}
