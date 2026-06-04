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
		EventOutput{Output: Output{Text: "test", IsErr: false}},
		EventError{Err: nil},
		EventCompleted{},
		EventClarifyingQuestions{Questions: []string{"Q?"}, AnswersCh: answersCh},
		EventPRDReview{PRD: &prd.PRD{}},
		EventPRDRevising{},
		EventCleanupStarted{},
		EventCleanupCompleted{},
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
	EventPRDReview{}.isEvent()
	EventPRDRevising{}.isEvent()
	EventCleanupStarted{}.isEvent()
	EventCleanupCompleted{}.isEvent()
}

func TestAllEventIsEventMethodsIncludesClarifying(t *testing.T) {
	answersCh := make(chan []prompt.QuestionAnswer, 1)
	EventClarifyingQuestions{Questions: []string{"Q?"}, AnswersCh: answersCh}.isEvent()
}

func TestEventCleanupStartedPassTotal(t *testing.T) {
	ev := EventCleanupStarted{Pass: 2, Total: 3}
	if ev.Pass != 2 || ev.Total != 3 {
		t.Fatalf("Pass=%d Total=%d, want Pass=2 Total=3", ev.Pass, ev.Total)
	}
}

func TestEventCleanupCompletedPassTotal(t *testing.T) {
	ev := EventCleanupCompleted{Pass: 1, Total: 3}
	if ev.Pass != 1 || ev.Total != 3 {
		t.Fatalf("Pass=%d Total=%d, want Pass=1 Total=3", ev.Pass, ev.Total)
	}
}
