package workflow

import (
	"errors"
	"testing"

	"ralph/internal/shared/runstate"
	"ralph/internal/workflow/events"
)

func TestEventStatusPhase(t *testing.T) {
	tests := []struct {
		name       string
		event      events.Event
		wantStatus string
		wantPhase  string
	}{
		{name: "clarify", event: events.EventClarifyingQuestions{}, wantStatus: runstate.StatusWaitingClarify, wantPhase: runstate.PhaseClarify},
		{name: "PRD review", event: events.EventPRDReview{}, wantStatus: runstate.StatusWaitingReview, wantPhase: runstate.PhaseReview},
		{name: "implementation review", event: events.EventImplementationReview{}, wantStatus: runstate.StatusWaitingImplReview, wantPhase: runstate.PhaseImplement},
		{name: "followup", event: events.EventStoryStarted{}, wantStatus: runstate.StatusImplementing, wantPhase: runstate.PhaseImplement},
		{name: "cleanup", event: events.EventCleanupStarted{}, wantStatus: runstate.StatusImplementing, wantPhase: runstate.PhaseCleanup},
		{name: "complete", event: events.EventCompleted{}, wantStatus: runstate.StatusCompleted, wantPhase: runstate.PhaseCompleted},
		{name: "failed", event: events.EventError{Err: errors.New("boom")}, wantStatus: runstate.StatusFailed, wantPhase: runstate.PhaseFailed},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotStatus, gotPhase := EventStatusPhase(tc.event)
			if gotStatus != tc.wantStatus || gotPhase != tc.wantPhase {
				t.Fatalf("EventStatusPhase() = (%q, %q), want (%q, %q)", gotStatus, gotPhase, tc.wantStatus, tc.wantPhase)
			}
		})
	}
}

func TestEventCheckpoint(t *testing.T) {
	tests := []struct {
		name  string
		event events.Event
		want  string
	}{
		{name: "PRD review", event: events.EventPRDReview{}, want: runstate.CheckpointPRDReview},
		{name: "implementation review", event: events.EventImplementationReview{}, want: runstate.CheckpointImplReview},
		{name: "complete", event: events.EventCompleted{}, want: runstate.CheckpointComplete},
		{name: "clarify", event: events.EventClarifyingQuestions{}, want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := EventCheckpoint(tc.event); got != tc.want {
				t.Fatalf("EventCheckpoint() = %q, want %q", got, tc.want)
			}
		})
	}
}
