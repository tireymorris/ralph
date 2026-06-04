package workflow

const (
	CheckpointImplReview = "impl_review"
	StopReasonDuplicateFindings = "duplicate_findings"
)

type ReviewLoopUpdate struct {
	Checkpoint               string
	ReviewIteration          int
	ReviewFingerprint        string
	ReviewElapsedMs          int64
	StopReason               string
	LastReviewTranscriptPath string
}

type ReviewLoopUpdater interface {
	Snapshot() (iteration int, fingerprint string, elapsedMs int64)
	Apply(u ReviewLoopUpdate) error
}
