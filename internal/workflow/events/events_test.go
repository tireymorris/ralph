package events

import (
	"testing"

	"ralph/internal/prd"
	"ralph/internal/prompt"
)

func TestEventTypes(t *testing.T) {
	answersCh := make(chan []prompt.QuestionAnswer, 1)
	evs := []Event{
		EventPRDGenerating{},
		EventPRDGenerated{PRD: &prd.PRD{}},
		EventPRDLoaded{PRD: &prd.PRD{}},
		EventStoryStarted{Story: &prd.Story{}, Iteration: 1},
		EventStoryCompleted{Story: &prd.Story{}, Success: true},
		EventOutput{Output: Output{Text: "test", IsErr: false}},
		EventError{Err: nil},
		EventCompleted{},
		EventFailed{FailedStories: nil},
		EventClarifyingQuestions{Questions: []string{"Q?"}, AnswersCh: answersCh},
		EventPRDReview{PRD: &prd.PRD{}},
	}

	for _, e := range evs {
		e.isEvent()
	}
}

func TestAllEventIsEventMethods(t *testing.T) {
	EventPRDGenerating{}.isEvent()
	EventPRDGenerated{}.isEvent()
	EventPRDLoaded{}.isEvent()
	EventStoryStarted{}.isEvent()
	EventStoryCompleted{}.isEvent()
	EventOutput{}.isEvent()
	EventError{}.isEvent()
	EventCompleted{}.isEvent()
	EventFailed{}.isEvent()
	EventPRDReview{}.isEvent()
}

func TestAllEventIsEventMethodsIncludesClarifying(t *testing.T) {
	answersCh := make(chan []prompt.QuestionAnswer, 1)
	EventClarifyingQuestions{Questions: []string{"Q?"}, AnswersCh: answersCh}.isEvent()
}
