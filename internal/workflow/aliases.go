package workflow

import "ralph/internal/workflow/events"

type (
	Event                    = events.Event
	Output                   = events.Output
	EventPRDGenerating       = events.EventPRDGenerating
	EventPRDGenerated        = events.EventPRDGenerated
	EventPRDLoaded           = events.EventPRDLoaded
	EventStoryStarted        = events.EventStoryStarted
	EventStoryCompleted      = events.EventStoryCompleted
	EventOutput              = events.EventOutput
	EventError               = events.EventError
	EventCompleted           = events.EventCompleted
	EventFailed              = events.EventFailed
	EventClarifyingQuestions = events.EventClarifyingQuestions
	EventPRDReview           = events.EventPRDReview
)
