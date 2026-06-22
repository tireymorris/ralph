package events

import (
	"encoding/json"
	"fmt"
)

type eventEnvelope struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

func MarshalEventEnvelope(ev Event) ([]byte, error) {
	typeName, payload, err := eventEnvelopeParts(ev)
	if err != nil {
		return nil, err
	}
	return json.Marshal(eventEnvelope{Type: typeName, Payload: payload})
}

func eventEnvelopeParts(ev Event) (string, any, error) {
	switch e := ev.(type) {
	case EventOutput:
		return "EventOutput", e.Output, nil
	case EventPRDGenerating:
		return "EventPRDGenerating", struct{}{}, nil
	case EventPRDGenerated:
		return "EventPRDGenerated", e.PRD, nil
	case EventPRDLoaded:
		return "EventPRDLoaded", e.PRD, nil
	case EventPRDRevising:
		return "EventPRDRevising", struct{}{}, nil
	case EventPRDReview:
		return "EventPRDReview", e.PRD, nil
	case EventStoryStarted:
		return "EventStoryStarted", e.Story, nil
	case EventStoryCompleted:
		return "EventStoryCompleted", struct {
			Story   any `json:"Story"`
			Success bool
		}{Story: e.Story, Success: e.Success}, nil
	case EventSliceStarted:
		return "EventSliceStarted", e, nil
	case EventSliceCompleted:
		return "EventSliceCompleted", e, nil
	case EventImplementationReviewStarted:
		return "EventImplementationReviewStarted", e, nil
	case EventImplementationReviewCompleted:
		return "EventImplementationReviewCompleted", e, nil
	case EventImplementationReview:
		return "EventImplementationReview", e, nil
	case EventRecoveryStarted:
		return "EventRecoveryStarted", e, nil
	case EventRecoveryCompleted:
		return "EventRecoveryCompleted", e, nil
	case EventCleanupStarted:
		return "EventCleanupStarted", e, nil
	case EventCleanupCompleted:
		return "EventCleanupCompleted", e, nil
	case EventCompleted:
		return "EventCompleted", struct{}{}, nil
	case EventError:
		msg := ""
		if e.Err != nil {
			msg = e.Err.Error()
		}
		return "EventError", map[string]string{"error": msg}, nil
	case EventClarifyingQuestions:
		return "EventClarifyingQuestions", struct {
			Questions []string `json:"Questions"`
		}{Questions: e.Questions}, nil
	default:
		return "", nil, fmt.Errorf("unknown event type %T", ev)
	}
}
