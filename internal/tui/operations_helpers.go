package tui

import (
	"fmt"

	"ralph/internal/workflow/events"
)

func (om *OperationManager) sendErrorEvent(err error) {
	if err == nil {
		return
	}
	select {
	case om.eventsCh <- events.EventError{Err: err}:
	case <-om.ctx.Done():
	}
}

func (om *OperationManager) launchBackgroundTask(fn func()) {
	go fn()
}

func wrapClarifyPhaseError(err error) error {
	return fmt.Errorf("clarification phase: %w", err)
}
