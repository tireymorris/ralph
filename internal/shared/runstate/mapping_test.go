package runstate

import (
	"testing"

	"ralph/internal/shared/prd"
)

func TestCheckpointStatusPhase(t *testing.T) {
	tests := []struct {
		name       string
		checkpoint string
		prd        *prd.PRD
		wantStatus string
		wantPhase  string
	}{
		{name: "PRD review", checkpoint: CheckpointPRDReview, prd: completedPRD(), wantStatus: StatusWaitingReview, wantPhase: PhaseReview},
		{name: "implementation review", checkpoint: CheckpointImplReview, prd: completedPRD(), wantStatus: StatusWaitingImplReview, wantPhase: PhaseCleanup},
		{name: "followup", checkpoint: CheckpointFollowup, prd: completedPRD(), wantStatus: StatusImplementing, wantPhase: PhaseImplement},
		{name: "complete", checkpoint: CheckpointComplete, prd: incompletePRD(), wantStatus: StatusCompleted, wantPhase: PhaseCompleted},
		{name: "incomplete PRD", prd: incompletePRD(), wantStatus: StatusImplementing, wantPhase: PhaseImplement},
		{name: "cancelled", checkpoint: CheckpointCancelled, prd: incompletePRD(), wantStatus: StatusCancelled, wantPhase: PhaseCancelled},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotStatus, gotPhase := CheckpointStatusPhase(tc.checkpoint, tc.prd)
			if gotStatus != tc.wantStatus || gotPhase != tc.wantPhase {
				t.Fatalf("CheckpointStatusPhase() = (%q, %q), want (%q, %q)", gotStatus, gotPhase, tc.wantStatus, tc.wantPhase)
			}
		})
	}
}

func TestLocalPRDStatusPhase(t *testing.T) {
	tests := []struct {
		name       string
		checkpoint string
		prd        *prd.PRD
		wantStatus string
		wantPhase  string
	}{
		{name: "incomplete PRD", prd: incompletePRD(), wantStatus: "implementing", wantPhase: PhaseImplement},
		{name: "implementation review", checkpoint: CheckpointImplReview, prd: incompletePRD(), wantStatus: StatusWaitingImplReview, wantPhase: PhaseCleanup},
		{name: "generate missing stories", prd: &prd.PRD{}, wantStatus: "running", wantPhase: PhaseGenerate},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotStatus, gotPhase := LocalPRDStatusPhase(tc.prd, tc.checkpoint)
			if gotStatus != tc.wantStatus || gotPhase != tc.wantPhase {
				t.Fatalf("LocalPRDStatusPhase() = (%q, %q), want (%q, %q)", gotStatus, gotPhase, tc.wantStatus, tc.wantPhase)
			}
		})
	}
}

func incompletePRD() *prd.PRD {
	return &prd.PRD{Stories: []*prd.Story{{ID: "story-1", Title: "one"}}}
}

func completedPRD() *prd.PRD {
	return &prd.PRD{Stories: []*prd.Story{{ID: "story-1", Title: "one", Passes: true}}}
}
