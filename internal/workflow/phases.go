package workflow

// DocOrchestratorPhase documents the executor state machine:
// clarify → generate/load → review → implement → complete.
type DocOrchestratorPhase int

const (
	DocPhaseClarify DocOrchestratorPhase = iota
	DocPhaseGenerate
	DocPhaseReview
	DocPhaseImplement
	DocPhaseComplete
)

func (p DocOrchestratorPhase) String() string {
	switch p {
	case DocPhaseClarify:
		return "clarify"
	case DocPhaseGenerate:
		return "generate_or_load"
	case DocPhaseReview:
		return "prd_review"
	case DocPhaseImplement:
		return "implement"
	case DocPhaseComplete:
		return "complete"
	default:
		return "unknown"
	}
}
