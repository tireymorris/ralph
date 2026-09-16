package runner

import "github.com/tireymorris/ralph/internal/workflow/events"

func MarshalEventEnvelope(ev events.Event) ([]byte, error) {
	return events.MarshalEventEnvelope(ev)
}
