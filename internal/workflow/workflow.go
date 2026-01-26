package workflow

import (
	"context"
	"fmt"
	"strings"

	"ralph/internal/config"
	"ralph/internal/constants"
	"ralph/internal/logger"
	"ralph/internal/prd"
	"ralph/internal/prompt"
	"ralph/internal/runner"
)

type Output struct {
	Text    string
	IsErr   bool
	Verbose bool
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

type Executor struct {
	cfg      *config.Config
	eventsCh chan Event
	runner   runner.RunnerInterface
}

func NewExecutor(cfg *config.Config, eventsCh chan Event) *Executor {
	return &Executor{
		cfg:      cfg,
		eventsCh: eventsCh,
		runner:   runner.New(cfg),
	}
}

func NewExecutorWithRunner(cfg *config.Config, eventsCh chan Event, r runner.RunnerInterface) *Executor {
	return &Executor{
		cfg:      cfg,
		eventsCh: eventsCh,
		runner:   r,
	}
}

func (e *Executor) RunGenerate(ctx context.Context, userPrompt string) (*prd.PRD, error) {
	logger.Debug("generating PRD", "prompt_length", len(userPrompt))
	e.emit(EventPRDGenerating{})
	e.emit(EventOutput{Output{Text: "Analyzing codebase and generating PRD..."}})

	outputCh := make(chan runner.OutputLine, constants.EventChannelBuffer)
	go e.forwardOutput(outputCh)

	prdPrompt := prompt.PRDGeneration(userPrompt, e.cfg.PRDFile, "feature")
	err := e.runner.Run(ctx, prdPrompt, outputCh)
	close(outputCh)

	if err != nil {
		logger.Error("PRD generation failed", "error", err)
		e.emit(EventError{Err: fmt.Errorf("PRD generation failed with model %s: %w", e.cfg.Model, err)})
		return nil, fmt.Errorf("PRD generation failed with model %s: %w", e.cfg.Model, err)
	}

	p, err := prd.Load(e.cfg)
	if err != nil {
		if isJSONParseError(err) {
			logger.Warn("PRD has JSON syntax error, attempting repair", "error", err)
			p, err = e.repairPRD(ctx, err)
			if err != nil {
				logger.Error("failed to repair PRD", "error", err)
				e.emit(EventError{Err: fmt.Errorf("PRD repair failed for %s: %w", e.cfg.PRDFile, err)})
				return nil, fmt.Errorf("PRD repair failed for %s: %w", e.cfg.PRDFile, err)
			}
		} else {
			logger.Error("failed to load generated PRD", "error", err)
			e.emit(EventError{Err: fmt.Errorf("failed to load generated PRD %s: %w", e.cfg.PRDFile, err)})
			return nil, fmt.Errorf("failed to load generated PRD %s: %w", e.cfg.PRDFile, err)
		}
	}

	logger.Debug("PRD generated", "project", p.ProjectName, "stories", len(p.Stories))
	e.emit(EventPRDGenerated{PRD: p})
	return p, nil
}

func (e *Executor) RunLoad(ctx context.Context) (*prd.PRD, error) {
	p, err := prd.Load(e.cfg)
	if err != nil {
		if isJSONParseError(err) {
			logger.Warn("PRD has JSON syntax error, attempting repair", "error", err)
			p, err = e.repairPRD(ctx, err)
			if err != nil {
				e.emit(EventError{Err: fmt.Errorf("PRD repair failed for %s: %w", e.cfg.PRDFile, err)})
				return nil, fmt.Errorf("PRD repair failed for %s: %w", e.cfg.PRDFile, err)
			}
		} else {
			e.emit(EventError{Err: fmt.Errorf("failed to load PRD %s: %w", e.cfg.PRDFile, err)})
			return nil, fmt.Errorf("failed to load PRD %s: %w", e.cfg.PRDFile, err)
		}
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

	iteration := 0

	for {
		select {
		case <-ctx.Done():
			logger.Debug("context cancelled")
			return ctx.Err()
		default:
		}

		p, err := prd.Load(e.cfg)
		if err != nil {
			logger.Error("failed to reload PRD", "error", err)
			wrappedErr := fmt.Errorf("failed to reload PRD %s: %w", e.cfg.PRDFile, err)
			e.emit(EventError{Err: fmt.Errorf("cannot continue without PRD: %w", wrappedErr)})
			return wrappedErr
		}

		if p.AllCompleted() {
			logger.Info("all stories completed successfully")
			prd.Delete(e.cfg)
			e.emit(EventCompleted{})
			return nil
		}

		next := p.NextPendingStory(e.cfg.RetryAttempts)
		if next == nil {
			failed := p.FailedStories(e.cfg.RetryAttempts)
			logger.Error("all remaining stories have failed", "failed_count", len(failed))
			e.emit(EventFailed{FailedStories: failed})
			return fmt.Errorf("all remaining stories have failed (%d stories)", len(failed))
		}

		iteration++
		if iteration > e.cfg.MaxIterations {
			logger.Error("max iterations reached", "iterations", iteration, "max", e.cfg.MaxIterations)
			e.emit(EventFailed{FailedStories: p.FailedStories(e.cfg.RetryAttempts)})
			return fmt.Errorf("max iterations (%d) reached after %d iterations", e.cfg.MaxIterations, iteration)
		}

		logger.Debug("starting story",
			"story_id", next.ID,
			"title", next.Title,
			"iteration", iteration,
			"retry_count", next.RetryCount)

		e.emit(EventStoryStarted{Story: next, Iteration: iteration})

		outputCh := make(chan runner.OutputLine, constants.EventChannelBuffer)
		go e.forwardOutput(outputCh)

		storyPrompt := prompt.StoryImplementation(
			next.ID,
			next.Title,
			next.Description,
			next.AcceptanceCriteria,
			p.TestSpec,
			p.Context,
			e.cfg.PRDFile,
			iteration,
			p.CompletedCount(),
			len(p.Stories),
		)

		err = e.runner.Run(ctx, storyPrompt, outputCh)
		close(outputCh)

		if err != nil {
			logger.Debug("AI runner returned error", "story_id", next.ID, "model", e.cfg.Model, "error", err)
		}

		updatedPRD, loadErr := prd.Load(e.cfg)
		if loadErr != nil {
			if isJSONParseError(loadErr) {
				logger.Warn("PRD corrupted during story, attempting repair", "error", loadErr, "story_id", next.ID)
				updatedPRD, loadErr = e.repairPRD(ctx, loadErr)
			}
			if loadErr != nil {
				logger.Error("failed to reload PRD after story, cannot continue", "error", loadErr, "story_id", next.ID)
				wrappedErr := fmt.Errorf("failed to reload PRD %s after story %s: %w", e.cfg.PRDFile, next.ID, loadErr)
				e.emit(EventError{Err: wrappedErr})
				return wrappedErr
			}
		}

		// Check for version conflicts (unexpected jumps indicate concurrent modification)
		if p.Version > 0 && updatedPRD.Version > p.Version+1 {
			logger.Warn("PRD version jumped unexpectedly",
				"previous", p.Version,
				"current", updatedPRD.Version,
				"expected", p.Version+1,
				"story_id", next.ID)
			e.emit(EventOutput{Output{Text: fmt.Sprintf(
				"Warning: PRD was modified externally (version %d → %d)", p.Version, updatedPRD.Version)}})
		}

		updatedStory := updatedPRD.GetStory(next.ID)
		if updatedStory != nil && updatedStory.Passes {
			logger.Debug("story marked as completed", "story_id", next.ID)
			e.emit(EventStoryCompleted{Story: updatedStory, Success: true})
		} else {
			logger.Debug("story not completed", "story_id", next.ID)
			if updatedStory != nil && updatedStory.RetryCount == next.RetryCount {
				updatedStory.RetryCount++
				if saveErr := prd.Save(e.cfg, updatedPRD); saveErr != nil {
					logger.Warn("failed to save retry count", "error", saveErr, "story_id", next.ID)
				}
			}
			e.emit(EventStoryCompleted{Story: next, Success: false})
		}

		p = updatedPRD
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
	for line := range outputCh {
		e.emit(EventOutput{Output{Text: line.Text, IsErr: line.IsErr, Verbose: line.Verbose}})
	}
}

func isJSONParseError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "invalid character") ||
		strings.Contains(errStr, "unexpected end of JSON") ||
		strings.Contains(errStr, "cannot unmarshal")
}

func (e *Executor) repairPRD(ctx context.Context, parseErr error) (*prd.PRD, error) {
	prdPath := e.cfg.PRDPath()

	for attempt := 1; attempt <= constants.MaxJSONRepairAttempts; attempt++ {
		logger.Debug("attempting to repair PRD JSON", "attempt", attempt, "file", prdPath, "error", parseErr.Error())
		e.emit(EventOutput{Output{Text: fmt.Sprintf("Attempting to repair malformed JSON in %s (attempt %d)...", prdPath, attempt)}})

		outputCh := make(chan runner.OutputLine, constants.EventChannelBuffer)
		go e.forwardOutput(outputCh)

		repairPrompt := prompt.JSONRepair(e.cfg.PRDFile, parseErr.Error())
		err := e.runner.Run(ctx, repairPrompt, outputCh)
		close(outputCh)

		if err != nil {
			logger.Warn("repair attempt failed", "attempt", attempt, "file", prdPath, "error", err)
			continue
		}

		p, loadErr := prd.Load(e.cfg)
		if loadErr == nil {
			logger.Debug("PRD JSON repaired successfully", "file", prdPath)
			e.emit(EventOutput{Output{Text: fmt.Sprintf("JSON repaired successfully in %s", prdPath)}})
			return p, nil
		}

		parseErr = loadErr
	}

	return nil, fmt.Errorf("failed to repair PRD JSON in %s after %d attempts: %w", prdPath, constants.MaxJSONRepairAttempts, parseErr)
}
