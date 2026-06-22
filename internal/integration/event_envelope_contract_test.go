package integration

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	webrunner "ralph/internal/web/runner"
	"ralph/internal/workflow/events"
)

func TestEventEnvelopeRunnerMatchesWorkflowEvents(t *testing.T) {
	cases := []events.Event{
		events.EventOutput{Output: events.Output{Text: "hello\nworld"}},
		events.EventImplementationReviewStarted{Iteration: 2},
		events.EventRecoveryCompleted{Reason: "test_gate", Attempt: 1, Success: true},
		events.EventSliceCompleted{StoryID: "s1", SliceID: "sl1"},
		events.EventError{Err: errors.New("boom")},
	}
	for _, ev := range cases {
		t.Run(fmt.Sprintf("%T", ev), func(t *testing.T) {
			fromRunner, err := webrunner.MarshalEventEnvelope(ev)
			if err != nil {
				t.Fatalf("runner.MarshalEventEnvelope() error = %v", err)
			}
			fromEvents, err := events.MarshalEventEnvelope(ev)
			if err != nil {
				t.Fatalf("events.MarshalEventEnvelope() error = %v", err)
			}
			if !bytes.Equal(fromRunner, fromEvents) {
				t.Fatalf("envelope mismatch:\n  runner: %s\n  events: %s", fromRunner, fromEvents)
			}
		})
	}
}
