package runstate

import (
	"testing"

	"ralph/internal/shared/prd"
)

func TestCheckpointPhase(t *testing.T) {
	tests := []struct {
		name       string
		checkpoint string
		prd        *prd.PRD
		want       string
	}{
		{name: "PRD review", checkpoint: CheckpointPRDReview, prd: completedPRD(), want: PhaseReview},
		{name: "implementation review", checkpoint: CheckpointImplReview, prd: completedPRD(), want: PhaseImplementationReview},
		{name: "followup", checkpoint: CheckpointFollowup, prd: completedPRD(), want: PhaseFollowup},
		{name: "complete", checkpoint: CheckpointComplete, prd: incompletePRD(), want: PhaseCompleted},
		{name: "incomplete PRD", prd: incompletePRD(), want: PhaseImplement},
		{name: "cancelled", checkpoint: CheckpointCancelled, prd: incompletePRD(), want: PhaseCancelled},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := CheckpointPhase(tc.checkpoint, tc.prd); got != tc.want {
				t.Fatalf("CheckpointPhase() = %q, want %q", got, tc.want)
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
		{name: "implementation review", checkpoint: CheckpointImplReview, prd: incompletePRD(), wantStatus: StatusWaitingImplReview, wantPhase: PhaseImplement},
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
