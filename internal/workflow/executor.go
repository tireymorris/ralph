package workflow

import (
	"context"
	"fmt"

	"ralph/internal/shared/config"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/logger"
	"ralph/internal/shared/runner"
)

// Executor orchestrates PRD generation, clarification, and story implementation.
type Executor struct {
	cfg      *config.Config
	eventsCh chan Event
	runner   runner.RunnerInterface
	store    PRDStore

	runID                  string
	reviewLoop             ReviewLoopUpdater
	reviewIteration        int
	reviewFingerprint      string
	reviewElapsedMs        int64
	reviewChangedFilesHash string
}

func NewExecutor(cfg *config.Config, eventsCh chan Event) *Executor {
	return &Executor{
		cfg:      cfg,
		eventsCh: eventsCh,
		runner:   runner.New(cfg),
		store:    defaultPRDStore{},
	}
}

func NewExecutorWithRunner(cfg *config.Config, eventsCh chan Event, r runner.RunnerInterface) *Executor {
	return NewExecutorWithRunnerAndStore(cfg, eventsCh, r, nil)
}

// NewExecutorWithRunnerAndStore injects a runner and optional store for tests.
func NewExecutorWithRunnerAndStore(cfg *config.Config, eventsCh chan Event, r runner.RunnerInterface, store PRDStore) *Executor {
	if store == nil {
		store = defaultPRDStore{}
	}
	return &Executor{
		cfg:      cfg,
		eventsCh: eventsCh,
		runner:   r,
		store:    store,
	}
}

func (e *Executor) emit(event Event) {
	if e.eventsCh != nil {
		select {
		case e.eventsCh <- event:
		default:
			logger.Warn("event channel full, dropping event", "event_type", fmt.Sprintf("%T", event))
		}
	}
}

func (e *Executor) forwardOutput(outputCh <-chan runner.OutputLine) {
	NewOutputForwarder(e.emit).Forward(outputCh)
}

func (e *Executor) RunPrompt(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
	return e.runner.Run(ctx, prompt, outputCh)
}

func (e *Executor) runWithForwardedOutput(ctx context.Context, prompt string) error {
	outputCh := make(chan runner.OutputLine, constants.EventChannelBuffer)
	done := make(chan struct{})
	go func() {
		e.forwardOutput(outputCh)
		close(done)
	}()
	runErr := e.runner.Run(ctx, prompt, outputCh)
	close(outputCh)
	<-done
	return runErr
}
