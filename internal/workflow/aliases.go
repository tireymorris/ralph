package workflow

import "ralph/internal/workflow/events"

type (
	Event                              = events.Event
	Output                             = events.Output
	EventPRDGenerating                 = events.EventPRDGenerating
	EventPRDGenerated                  = events.EventPRDGenerated
	EventPRDLoaded                     = events.EventPRDLoaded
	EventStoryStarted                  = events.EventStoryStarted
	EventStoryCompleted                = events.EventStoryCompleted
	EventOutput                        = events.EventOutput
	EventError                         = events.EventError
	EventCompleted                     = events.EventCompleted
	EventClarifyingQuestions           = events.EventClarifyingQuestions
	EventPRDReview                     = events.EventPRDReview
	EventPRDRevising                   = events.EventPRDRevising
	EventCleanupStarted                = events.EventCleanupStarted
	EventCleanupCompleted              = events.EventCleanupCompleted
	EventImplementationReviewStarted   = events.EventImplementationReviewStarted
	EventImplementationReviewCompleted = events.EventImplementationReviewCompleted
	EventImplementationReview          = events.EventImplementationReview
	EventRecoveryStarted               = events.EventRecoveryStarted
	EventRecoveryCompleted             = events.EventRecoveryCompleted
	ImplementationFinding              = events.ImplementationFinding
)
