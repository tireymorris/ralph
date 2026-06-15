package events

import (
	"testing"

	"ralph/internal/prompt"
	"ralph/internal/shared/prd"
)

func TestEventTypes(t *testing.T) {
	answersCh := make(chan []prompt.QuestionAnswer, 1)
	evs := []Event{
		EventPRDGenerating{},
		EventPRDGenerated{PRD: &prd.PRD{}},
		EventPRDLoaded{PRD: &prd.PRD{}},
		EventStoryStarted{Story: &prd.Story{}},
		EventStoryCompleted{Story: &prd.Story{}, Success: true},
		EventSliceStarted{StoryID: "story-1", SliceID: "slice-1"},
		EventSliceCompleted{StoryID: "story-1", SliceID: "slice-1"},
		EventOutput{Output: Output{Text: "test", IsErr: false}},
		EventError{Err: nil},
		EventCompleted{},
		EventClarifyingQuestions{Questions: []string{"Q?"}, AnswersCh: answersCh},
		EventPRDReview{PRD: &prd.PRD{}},
		EventPRDRevising{},
		EventCleanupStarted{},
		EventCleanupCompleted{},
		EventImplementationReviewStarted{Iteration: 1},
		EventImplementationReviewCompleted{Iteration: 1, Clean: true},
		EventImplementationReview{Findings: []ImplementationFinding{{ID: "x", Summary: "s"}}},
		EventRecoveryStarted{Reason: "review_findings", Attempt: 1, Max: 2},
		EventRecoveryCompleted{Reason: "review_findings", Attempt: 1, Success: true},
	}

	for _, e := range evs {
		e.isEvent()
	}
}

func TestAllEventIsEventMethods(t *testing.T) {
	evs := []Event{
		EventPRDGenerating{},
		EventPRDGenerated{},
		EventPRDLoaded{},
		EventStoryStarted{},
		EventStoryCompleted{},
		EventSliceStarted{},
		EventSliceCompleted{},
		EventOutput{},
		EventError{},
		EventCompleted{},
		EventPRDReview{},
		EventPRDRevising{},
		EventCleanupStarted{},
		EventCleanupCompleted{},
		EventImplementationReviewStarted{},
		EventImplementationReviewCompleted{},
		EventImplementationReview{},
		EventRecoveryStarted{},
		EventRecoveryCompleted{},
	}
	for _, e := range evs {
		e.isEvent()
	}
}

func TestAllEventIsEventMethodsIncludesClarifying(t *testing.T) {
	answersCh := make(chan []prompt.QuestionAnswer, 1)
	EventClarifyingQuestions{Questions: []string{"Q?"}, AnswersCh: answersCh}.isEvent()
}
