// Package events defines workflow notifications emitted to the UI and CLI.
// It is a leaf package: orchestration lives in workflow.Executor.
package events

import (
	"ralph/internal/prd"
	"ralph/internal/prompt"
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
	Story     *prd.Story
	Iteration int
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

type EventFailed struct {
	FailedStories []*prd.Story
}

func (EventFailed) isEvent() {}

// EventClarifyingQuestions is emitted when the AI has generated clarifying
// questions. The consumer (TUI or CLI) should collect answers and send them
// back via the AnswersCh channel.
type EventClarifyingQuestions struct {
	Questions []string
	AnswersCh chan<- []prompt.QuestionAnswer
}

func (EventClarifyingQuestions) isEvent() {}

// EventPRDReview is emitted after PRD generation to signal the consumer should
// prompt the user to review before proceeding to implementation.
type EventPRDReview struct {
	PRD *prd.PRD
}

func (EventPRDReview) isEvent() {}
