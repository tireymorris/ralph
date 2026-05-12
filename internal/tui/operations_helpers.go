package tui

import (
	"fmt"

	"ralph/internal/workflow/events"
)

func (om *OperationManager) emitError(err error) {
	if err == nil {
		return
	}
	select {
	case om.eventsCh <- events.EventError{Err: err}:
	case <-om.ctx.Done():
	}
}

func (om *OperationManager) startBackground(fn func()) {
	go fn()
}

func clarifyPhaseError(err error) error {
	return fmt.Errorf("clarification phase: %w", err)
}
