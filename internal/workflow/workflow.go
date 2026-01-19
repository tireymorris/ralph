package workflow

import (
	"context"
	"fmt"

	"ralph/internal/config"
	"ralph/internal/git"
	"ralph/internal/logger"
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
	logger.Debug("generating PRD", "prompt_length", len(prompt))
	e.emit(EventPRDGenerating{})
	e.emit(EventOutput{Output{Text: "Analyzing codebase and generating PRD...", IsErr: false}})

	outputCh := make(chan runner.OutputLine, 100)

	go e.forwardOutput(outputCh)

	p, err := e.generator.Generate(ctx, prompt, outputCh)
	close(outputCh)

	if err != nil {
		logger.Error("PRD generation failed", "error", err)
		e.emit(EventError{Err: err})
		return nil, err
	}

	logger.Debug("PRD generated", "project", p.ProjectName, "stories", len(p.Stories))

	if err := e.storage.Save(p); err != nil {
		logger.Error("failed to save PRD", "error", err)
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
	logger.Debug("starting implementation",
		"project", p.ProjectName,
		"branch", p.BranchName,
		"total_stories", len(p.Stories),
		"completed", p.CompletedCount())

	if p.BranchName != "" {
		logger.Debug("creating branch", "branch", p.BranchName)
		if err := e.git.CreateBranch(p.BranchName); err != nil {
			logger.Warn("failed to create branch", "branch", p.BranchName, "error", err)
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
			logger.Debug("context cancelled")
			return ctx.Err()
		default:
		}

		if p.AllCompleted() {
			logger.Info("all stories completed successfully")
			e.storage.Delete()
			e.emit(EventCompleted{})
			return nil
		}

		next := p.NextPendingStory(e.cfg.RetryAttempts)
		if next == nil {
			failed := p.FailedStories(e.cfg.RetryAttempts)
			logger.Error("all remaining stories have failed", "failed_count", len(failed))
			e.emit(EventFailed{FailedStories: failed})
			return fmt.Errorf("all remaining stories have failed")
		}

		iteration++
		if iteration > e.cfg.MaxIterations {
			logger.Error("max iterations reached", "iterations", iteration, "max", e.cfg.MaxIterations)
			e.emit(EventFailed{FailedStories: p.FailedStories(e.cfg.RetryAttempts)})
			return fmt.Errorf("max iterations (%d) reached", e.cfg.MaxIterations)
		}

		logger.Debug("starting story",
			"story_id", next.ID,
			"title", next.Title,
			"iteration", iteration,
			"retry_count", next.RetryCount)

		e.emit(EventStoryStarted{Story: next, Iteration: iteration})

		outputCh := make(chan runner.OutputLine, 100)
		go e.forwardOutput(outputCh)

		success, err := e.implementer.Implement(ctx, next, iteration, p, outputCh)
		close(outputCh)

		if err != nil {
			logger.Debug("story implementation error", "story_id", next.ID, "error", err)
			next.RetryCount++
			e.emit(EventStoryCompleted{Story: next, Success: false})
		} else if success {
			logger.Debug("story completed successfully", "story_id", next.ID)
			next.Passes = true
			e.emit(EventStoryCompleted{Story: next, Success: true})
		} else {
			logger.Debug("story failed", "story_id", next.ID, "retry_count", next.RetryCount+1)
			next.RetryCount++
			e.emit(EventStoryCompleted{Story: next, Success: false})
		}

		if err := e.storage.Save(p); err != nil {
			logger.Warn("failed to save state", "error", err)
			e.emit(EventOutput{Output{
				Text:  fmt.Sprintf("Warning: failed to save state: %v", err),
				IsErr: true,
			}})
		}
	}
}

func (e *Executor) emit(event Event) {
	if e.eventsCh != nil {
		// Use non-blocking send to prevent deadlock if event channel is full.
		// In practice, the channel is buffered (100) and consumers should keep up,
		// but this prevents blocking the workflow if they don't.
		select {
		case e.eventsCh <- event:
			// Event sent successfully
		default:
			// Channel full, log and drop the event to prevent deadlock
			logger.Warn("event channel full, dropping event", "event_type", fmt.Sprintf("%T", event))
		}
	}
}

func (e *Executor) forwardOutput(outputCh <-chan runner.OutputLine) {
	for line := range outputCh {
		e.emit(EventOutput{Output{Text: line.Text, IsErr: line.IsErr}})
	}
}
