package events

import (
	"encoding/json"
	"testing"
)

func assertEnvelopeType(t *testing.T, ev Event, wantType string) {
	t.Helper()
	data, err := MarshalEventEnvelope(ev)
	if err != nil {
		t.Fatalf("MarshalEventEnvelope() error = %v", err)
	}
	var env eventEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if env.Type != wantType {
		t.Errorf("Type = %q, want %q", env.Type, wantType)
	}
}

func TestMarshalEventEnvelope_ImplementationReviewEvents(t *testing.T) {
	cases := []struct {
		name     string
		ev       Event
		wantType string
	}{
		{name: "started", ev: EventImplementationReviewStarted{Iteration: 1}, wantType: "EventImplementationReviewStarted"},
		{name: "completed", ev: EventImplementationReviewCompleted{Iteration: 1, Clean: true}, wantType: "EventImplementationReviewCompleted"},
		{name: "findings", ev: EventImplementationReview{Findings: []ImplementationFinding{{ID: "a", Summary: "s"}}}, wantType: "EventImplementationReview"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertEnvelopeType(t, tc.ev, tc.wantType)
		})
	}
}

func TestMarshalEventEnvelope_RecoveryEvents(t *testing.T) {
	cases := []struct {
		name     string
		ev       Event
		wantType string
	}{
		{name: "started", ev: EventRecoveryStarted{Reason: "test_gate", Attempt: 1, Max: 2}, wantType: "EventRecoveryStarted"},
		{name: "completed", ev: EventRecoveryCompleted{Reason: "test_gate", Attempt: 1, Success: true}, wantType: "EventRecoveryCompleted"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertEnvelopeType(t, tc.ev, tc.wantType)
		})
	}
}

func TestMarshalEventEnvelope_CleanupEvents(t *testing.T) {
	cases := []struct {
		name     string
		ev       Event
		wantType string
	}{
		{name: "started", ev: EventCleanupStarted{}, wantType: "EventCleanupStarted"},
		{name: "completed", ev: EventCleanupCompleted{}, wantType: "EventCleanupCompleted"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertEnvelopeType(t, tc.ev, tc.wantType)
		})
	}
}

func TestMarshalEventEnvelope_SliceEvents(t *testing.T) {
	cases := []struct {
		name     string
		ev       Event
		wantType string
	}{
		{name: "started", ev: EventSliceStarted{StoryID: "story-1", SliceID: "slice-1"}, wantType: "EventSliceStarted"},
		{name: "completed", ev: EventSliceCompleted{StoryID: "story-1", SliceID: "slice-1"}, wantType: "EventSliceCompleted"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertEnvelopeType(t, tc.ev, tc.wantType)
		})
	}
}
