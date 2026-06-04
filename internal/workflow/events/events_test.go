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

func TestCleanupPassProgress(t *testing.T) {
	cases := []struct {
		name  string
		pass  int
		total int
	}{
		{name: "started pass 2", pass: 2, total: 3},
		{name: "completed pass 1", pass: 1, total: 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			progress := CleanupPassProgress{Pass: tc.pass, Total: tc.total}
			if progress.Pass != tc.pass || progress.Total != tc.total {
				t.Fatalf("Pass=%d Total=%d, want Pass=%d Total=%d", progress.Pass, progress.Total, tc.pass, tc.total)
			}
		})
	}
}
