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

type Output struct {
	Text  string
	IsErr bool
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

type PRDGenerator interface {
	Generate(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) (*prd.PRD, error)
}

type StoryImplementer interface {
	Implement(ctx context.Context, story *prd.Story, iteration int, p *prd.PRD, outputCh chan<- runner.OutputLine) (bool, error)
}

type GitManager interface {
	CreateBranch(name string) error
}

type PRDStorage interface {
	Load() (*prd.PRD, error)
	Save(p *prd.PRD) error
	Delete() error
}

type defaultPRDStorage struct {
	cfg *config.Config
}

func (s *defaultPRDStorage) Load() (*prd.PRD, error) {
	return prd.Load(s.cfg)
}

func (s *defaultPRDStorage) Save(p *prd.PRD) error {
	return prd.Save(s.cfg, p)
}

func (s *defaultPRDStorage) Delete() error {
	return prd.Delete(s.cfg)
}

type Executor struct {
	cfg         *config.Config
	eventsCh    chan Event
	generator   PRDGenerator
	implementer StoryImplementer
	git         GitManager
	storage     PRDStorage
}

func NewExecutor(cfg *config.Config, eventsCh chan Event) *Executor {
	return &Executor{
		cfg:         cfg,
		eventsCh:    eventsCh,
		generator:   prd.NewGenerator(cfg),
		implementer: story.NewImplementer(cfg),
		git:         git.NewWithWorkDir(cfg.WorkDir),
		storage:     &defaultPRDStorage{cfg: cfg},
	}
}

func NewExecutorWithDeps(cfg *config.Config, eventsCh chan Event, gen PRDGenerator, impl StoryImplementer, g GitManager, storage PRDStorage) *Executor {
	return &Executor{
		cfg:         cfg,
		eventsCh:    eventsCh,
		generator:   gen,
		implementer: impl,
		git:         g,
		storage:     storage,
	}
}

func (e *Executor) RunGenerate(ctx context.Context, prompt string) (*prd.PRD, error) {
	e.emit(EventPRDGenerating{})

	outputCh := make(chan runner.OutputLine, 100)

	go e.forwardOutput(outputCh)

	p, err := e.generator.Generate(ctx, prompt, outputCh)
	close(outputCh)

	if err != nil {
		e.emit(EventError{Err: err})
		return nil, err
	}

	if err := e.storage.Save(p); err != nil {
		e.emit(EventError{Err: err})
		return nil, err
	}

	e.emit(EventPRDGenerated{PRD: p})
	return p, nil
}

func (e *Executor) RunLoad(ctx context.Context) (*prd.PRD, error) {
	p, err := e.storage.Load()
	if err != nil {
		e.emit(EventError{Err: err})
		return nil, err
	}

	e.emit(EventPRDLoaded{PRD: p})
	return p, nil
}

func (e *Executor) RunImplementation(ctx context.Context, p *prd.PRD) error {
	if p.BranchName != "" {
		if err := e.git.CreateBranch(p.BranchName); err != nil {
			e.emit(EventOutput{Output{
				Text:  fmt.Sprintf("Warning: failed to create branch: %v", err),
				IsErr: true,
			}})
		}
	}

	iteration := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if p.AllCompleted() {
			e.storage.Delete()
			e.emit(EventCompleted{})
			return nil
		}

		next := p.NextPendingStory(e.cfg.RetryAttempts)
		if next == nil {
			e.emit(EventFailed{FailedStories: p.FailedStories(e.cfg.RetryAttempts)})
			return fmt.Errorf("all remaining stories have failed")
		}

		iteration++
		if iteration > e.cfg.MaxIterations {
			e.emit(EventFailed{FailedStories: p.FailedStories(e.cfg.RetryAttempts)})
			return fmt.Errorf("max iterations (%d) reached", e.cfg.MaxIterations)
		}

		e.emit(EventStoryStarted{Story: next, Iteration: iteration})

		outputCh := make(chan runner.OutputLine, 100)
		go e.forwardOutput(outputCh)

		success, err := e.implementer.Implement(ctx, next, iteration, p, outputCh)
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

		if err := e.storage.Save(p); err != nil {
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
