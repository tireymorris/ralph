package runner

import (
	"encoding/json"
	"testing"

	"ralph/internal/workflow/events"
)

func TestMarshalEventEnvelope_CleanupStarted(t *testing.T) {
	data, err := MarshalEventEnvelope(events.EventCleanupStarted{})
	if err != nil {
		t.Fatalf("MarshalEventEnvelope() error = %v", err)
	}
	var env eventEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if env.Type != "EventCleanupStarted" {
		t.Errorf("Type = %q, want %q", env.Type, "EventCleanupStarted")
	}
}

func TestMarshalEventEnvelope_CleanupCompleted(t *testing.T) {
	data, err := MarshalEventEnvelope(events.EventCleanupCompleted{})
	if err != nil {
		t.Fatalf("MarshalEventEnvelope() error = %v", err)
	}
	var env eventEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if env.Type != "EventCleanupCompleted" {
		t.Errorf("Type = %q, want %q", env.Type, "EventCleanupCompleted")
	}
}
