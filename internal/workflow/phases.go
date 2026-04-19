package workflow

// DocOrchestratorPhase names high-level executor stages for navigation and logging.
// It documents the conceptual state machine; transitions are implemented by Executor methods.
//
// Transitions (happy path, non–dry-run):
//
//	DocPhaseClarify   → RunClarify (optional; may skip if no questions)
//	DocPhaseGenerate  → RunGenerateWithAnswers / RunGenerate / RunLoad (--resume)
//	DocPhaseReview    → EventPRDReview (consumer confirms or edits PRD)
//	DocPhaseImplement → RunImplementation (story loop until all pass)
//	DocPhaseComplete  → EventCompleted
//
// See RunClarify, RunGenerateWithAnswers, RunLoad, RunImplementation, and phase_*.go.
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
