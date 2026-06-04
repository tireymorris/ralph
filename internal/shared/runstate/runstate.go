package runstate

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
