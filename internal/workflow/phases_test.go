package workflow

import "testing"

func TestDocOrchestratorPhaseString(t *testing.T) {
	tests := []struct {
		p    DocOrchestratorPhase
		want string
	}{
		{DocPhaseClarify, "clarify"},
		{DocPhaseGenerate, "generate_or_load"},
		{DocPhaseReview, "prd_review"},
		{DocPhaseImplement, "implement"},
		{DocPhaseComplete, "complete"},
		{DocPhaseFailed, "failed"},
		{DocOrchestratorPhase(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.p.String(); got != tt.want {
			t.Errorf("DocOrchestratorPhase(%v).String() = %q, want %q", tt.p, got, tt.want)
		}
	}
}
