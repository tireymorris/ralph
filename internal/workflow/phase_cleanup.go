package workflow

import (
	"context"
	"fmt"

	"ralph/internal/prompt"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
)

func (e *Executor) RunCleanup(ctx context.Context, p *prd.PRD) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	e.emit(EventCleanupStarted{})

	cleanupPrompt := prompt.Cleanup(p.Context, e.cfg.PRDFile)

	outputCh := make(chan runner.OutputLine, constants.EventChannelBuffer)
	done := make(chan struct{})
	go func() {
		e.forwardOutput(outputCh)
		close(done)
	}()

	runErr := e.runner.Run(ctx, cleanupPrompt, outputCh)
	close(outputCh)
	<-done

	if runErr != nil {
		e.emit(EventError{Err: fmt.Errorf("cleanup failed: %w", runErr)})
		return runErr
	}

	e.emit(EventCleanupCompleted{})
	return nil
}
