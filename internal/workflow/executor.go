package workflow

import (
	"fmt"

	"ralph/internal/config"
	"ralph/internal/logger"
	"ralph/internal/runner"
)

// Executor orchestrates PRD generation, clarification, and story implementation.
type Executor struct {
	cfg      *config.Config
	eventsCh chan Event
	runner   runner.RunnerInterface
	store    PRDStore
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

// NewExecutorWithRunnerAndStore constructs an executor with a custom AI runner
// and optional PRD persistence. If store is nil, the default disk-backed store is used.
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
