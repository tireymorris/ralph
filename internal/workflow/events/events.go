// Package events defines workflow notifications emitted to UI and CLI consumers.
package events

import (
	"ralph/internal/prompt"
	"ralph/internal/shared/prd"
)

type Output struct {
	Text    string
	IsErr   bool
	Verbose bool
}

type Event interface {
	isEvent()
}

type EventPRDGenerating struct{}

func (EventPRDGenerating) isEvent() {}

type EventPRDGenerated struct {
	PRD *prd.PRD
}

func (EventPRDGenerated) isEvent() {}

type EventPRDLoaded struct {
	PRD *prd.PRD
}

func (EventPRDLoaded) isEvent() {}

type EventStoryStarted struct {
	Story *prd.Story
}

func (EventStoryStarted) isEvent() {}

type EventStoryCompleted struct {
	Story   *prd.Story
	Success bool
}

func (EventStoryCompleted) isEvent() {}

type EventOutput struct {
	Output
}

func (EventOutput) isEvent() {}

type EventError struct {
	Err error
}

func (EventError) isEvent() {}

type EventCompleted struct{}

func (EventCompleted) isEvent() {}

// EventClarifyingQuestions carries an answer channel the consumer must reply on.
type EventClarifyingQuestions struct {
	Questions []string
	AnswersCh chan<- []prompt.QuestionAnswer
}

func (EventClarifyingQuestions) isEvent() {}

// EventPRDReview asks the consumer to review before implementation proceeds.
type EventPRDReview struct {
	PRD *prd.PRD
}

func (EventPRDReview) isEvent() {}

type EventPRDRevising struct{}

func (EventPRDRevising) isEvent() {}

type EventCleanupStarted struct{}

func (EventCleanupStarted) isEvent() {}

type EventCleanupCompleted struct{}

func (EventCleanupCompleted) isEvent() {}

type ImplementationFinding struct {
	ID       string
	Category string
	Path     string
	Line     int
	Summary  string
}

type EventImplementationReviewStarted struct {
	Iteration int
}

func (EventImplementationReviewStarted) isEvent() {}

type EventImplementationReviewCompleted struct {
	Iteration int
	Clean     bool
}

func (EventImplementationReviewCompleted) isEvent() {}

type EventImplementationReview struct {
	Findings []ImplementationFinding
}

func (EventImplementationReview) isEvent() {}

type EventRecoveryStarted struct {
	Reason  string
	Attempt int
	Max     int
}

func (EventRecoveryStarted) isEvent() {}

type EventRecoveryCompleted struct {
	Reason  string
	Attempt int
	Success bool
}

func (EventRecoveryCompleted) isEvent() {}
