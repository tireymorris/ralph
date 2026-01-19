package workflow

import (
	"context"
	"fmt"

	"ralph/internal/config"
	"ralph/internal/git"
	"ralph/internal/prd"
	"ralph/internal/runner"
	"ralph/internal/story"
)

// Output represents a workflow output event
type Output struct {
	Text  string
	IsErr bool
}

// Event represents a workflow event
type Event interface {
	isEvent()
}

// EventPRDGenerating indicates PRD generation has started
type EventPRDGenerating struct{}

func (EventPRDGenerating) isEvent() {}

// EventPRDGenerated indicates PRD was successfully generated
type EventPRDGenerated struct {
	PRD *prd.PRD
}

func (EventPRDGenerated) isEvent() {}

// EventPRDLoaded indicates PRD was loaded from file
type EventPRDLoaded struct {
	PRD *prd.PRD
}

func (EventPRDLoaded) isEvent() {}

// EventStoryStarted indicates a story implementation has started
type EventStoryStarted struct {
	Story     *prd.Story
	Iteration int
}

func (EventStoryStarted) isEvent() {}

// EventStoryCompleted indicates a story was completed
type EventStoryCompleted struct {
	Story   *prd.Story
	Success bool
}

func (EventStoryCompleted) isEvent() {}

// EventOutput represents output from the underlying command
type EventOutput struct {
	Output
}

func (EventOutput) isEvent() {}

// EventError indicates an error occurred
type EventError struct {
	Err error
}

func (EventError) isEvent() {}

// EventCompleted indicates all stories are done
type EventCompleted struct{}

func (EventCompleted) isEvent() {}

// EventFailed indicates the workflow failed
type EventFailed struct {
	FailedStories []*prd.Story
}

func (EventFailed) isEvent() {}

// Executor runs the workflow
type Executor struct {
	cfg      *config.Config
	eventsCh chan Event
}

// NewExecutor creates a new workflow executor
func NewExecutor(cfg *config.Config, eventsCh chan Event) *Executor {
	return &Executor{
		cfg:      cfg,
		eventsCh: eventsCh,
	}
}

// RunGenerate generates a PRD from the prompt
func (e *Executor) RunGenerate(ctx context.Context, prompt string) (*prd.PRD, error) {
	e.emit(EventPRDGenerating{})

	gen := prd.NewGenerator(e.cfg)
	outputCh := make(chan runner.OutputLine, 100)

	go e.forwardOutput(outputCh)

	p, err := gen.Generate(ctx, prompt, outputCh)
	close(outputCh)

	if err != nil {
		e.emit(EventError{Err: err})
		return nil, err
	}

	if err := prd.Save(e.cfg, p); err != nil {
		e.emit(EventError{Err: err})
		return nil, err
	}

	e.emit(EventPRDGenerated{PRD: p})
	return p, nil
}

// RunLoad loads an existing PRD
func (e *Executor) RunLoad(ctx context.Context) (*prd.PRD, error) {
	p, err := prd.Load(e.cfg)
	if err != nil {
		e.emit(EventError{Err: err})
		return nil, err
	}

	e.emit(EventPRDLoaded{PRD: p})
	return p, nil
}

// RunImplementation implements all stories in the PRD
func (e *Executor) RunImplementation(ctx context.Context, p *prd.PRD) error {
	// Setup branch if specified
	if p.BranchName != "" {
		gitMgr := git.New()
		if err := gitMgr.CreateBranch(p.BranchName); err != nil {
			e.emit(EventOutput{Output{
				Text:  fmt.Sprintf("Warning: failed to create branch: %v", err),
				IsErr: true,
			}})
		}
	}

	impl := story.NewImplementer(e.cfg)
	iteration := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check if all done
		if p.AllCompleted() {
			prd.Delete(e.cfg)
			e.emit(EventCompleted{})
			return nil
		}

		// Get next story
		next := p.NextPendingStory(e.cfg.RetryAttempts)
		if next == nil {
			e.emit(EventFailed{FailedStories: p.FailedStories(e.cfg.RetryAttempts)})
			return fmt.Errorf("all remaining stories have failed")
		}

		// Check max iterations
		iteration++
		if iteration > e.cfg.MaxIterations {
			e.emit(EventFailed{FailedStories: p.FailedStories(e.cfg.RetryAttempts)})
			return fmt.Errorf("max iterations (%d) reached", e.cfg.MaxIterations)
		}

		// Start story
		e.emit(EventStoryStarted{Story: next, Iteration: iteration})

		// Run implementation
		outputCh := make(chan runner.OutputLine, 100)
		go e.forwardOutput(outputCh)

		success, err := impl.Implement(ctx, next, iteration, p, outputCh)
		close(outputCh)

		if err != nil {
			next.RetryCount++
			e.emit(EventStoryCompleted{Story: next, Success: false})
		} else if success {
			next.Passes = true
			e.emit(EventStoryCompleted{Story: next, Success: true})
		} else {
			next.RetryCount++
			e.emit(EventStoryCompleted{Story: next, Success: false})
		}

		// Save state
		if err := prd.Save(e.cfg, p); err != nil {
			e.emit(EventOutput{Output{
				Text:  fmt.Sprintf("Warning: failed to save state: %v", err),
				IsErr: true,
			}})
		}
	}
}

func (e *Executor) emit(event Event) {
	if e.eventsCh != nil {
		e.eventsCh <- event
	}
}

func (e *Executor) forwardOutput(outputCh <-chan runner.OutputLine) {
	for line := range outputCh {
		e.emit(EventOutput{Output{Text: line.Text, IsErr: line.IsErr}})
	}
}
