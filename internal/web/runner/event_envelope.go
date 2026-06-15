package runner

import (
	"encoding/json"
	"fmt"

	"ralph/internal/workflow/events"
)

type eventEnvelope struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

func MarshalEventEnvelope(ev events.Event) ([]byte, error) {
	typeName, payload, err := eventEnvelopeParts(ev)
	if err != nil {
		return nil, err
	}
	return json.Marshal(eventEnvelope{Type: typeName, Payload: payload})
}

func eventEnvelopeParts(ev events.Event) (string, any, error) {
	switch e := ev.(type) {
	case events.EventOutput:
		return "EventOutput", e.Output, nil
	case events.EventPRDGenerating:
		return "EventPRDGenerating", struct{}{}, nil
	case events.EventPRDGenerated:
		return "EventPRDGenerated", e.PRD, nil
	case events.EventPRDLoaded:
		return "EventPRDLoaded", e.PRD, nil
	case events.EventPRDRevising:
		return "EventPRDRevising", struct{}{}, nil
	case events.EventPRDReview:
		return "EventPRDReview", e.PRD, nil
	case events.EventStoryStarted:
		return "EventStoryStarted", e.Story, nil
	case events.EventStoryCompleted:
		return "EventStoryCompleted", struct {
			Story   any `json:"Story"`
			Success bool
		}{Story: e.Story, Success: e.Success}, nil
	case events.EventSliceStarted:
		return "EventSliceStarted", e, nil
	case events.EventSliceCompleted:
		return "EventSliceCompleted", e, nil
	case events.EventImplementationReviewStarted:
		return "EventImplementationReviewStarted", e, nil
	case events.EventImplementationReviewCompleted:
		return "EventImplementationReviewCompleted", e, nil
	case events.EventImplementationReview:
		return "EventImplementationReview", e, nil
	case events.EventCleanupStarted:
		return "EventCleanupStarted", e, nil
	case events.EventCleanupCompleted:
		return "EventCleanupCompleted", e, nil
	case events.EventCompleted:
		return "EventCompleted", struct{}{}, nil
	case events.EventError:
		msg := ""
		if e.Err != nil {
			msg = e.Err.Error()
		}
		return "EventError", map[string]string{"error": msg}, nil
	case events.EventClarifyingQuestions:
		return "EventClarifyingQuestions", struct {
			Questions []string `json:"Questions"`
		}{Questions: e.Questions}, nil
	default:
		return "", nil, fmt.Errorf("unknown event type %T", ev)
	}
}
