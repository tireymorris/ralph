package workflow

import (
	"context"
	"fmt"

	"ralph/internal/constants"
	"ralph/internal/logger"
	"ralph/internal/prd"
	"ralph/internal/prompt"
	"ralph/internal/runner"
)

func (e *Executor) RunImplementation(ctx context.Context, p *prd.PRD) error {
	logger.Debug("starting implementation",
		"project", p.ProjectName,
		"branch", p.BranchName,
		"total_stories", len(p.Stories),
		"completed", p.CompletedCount())

	for {
		select {
		case <-ctx.Done():
			logger.Debug("context cancelled")
			return ctx.Err()
		default:
		}

		p, err := e.store.Load(e.cfg)
		if err != nil {
			logger.Error("failed to reload PRD", "error", err)
			wrappedErr := fmt.Errorf("failed to reload PRD %s: %w", e.cfg.PRDFile, err)
			e.emit(EventError{Err: fmt.Errorf("cannot continue without PRD: %w", wrappedErr)})
			return wrappedErr
		}

		if p.AllCompleted() {
			logger.Info("all stories completed successfully")
			e.emit(EventCompleted{})
			return nil
		}

		ready := p.ReadyStories()
		if len(ready) == 0 {
			blocked := p.BlockedStories()
			if len(blocked) > 0 {
				logger.Warn("no ready stories, waiting for dependencies", "blocked_count", len(blocked))
			}
			continue
		}

		story := ready[0]

		logger.Debug("starting story",
			"story_id", story.ID,
			"title", story.Title)

		e.emit(EventStoryStarted{Story: story})

		outputCh := make(chan runner.OutputLine, constants.EventChannelBuffer)
		go e.forwardOutput(outputCh)

		storyPrompt := prompt.StoryImplementation(
			story.ID,
			story.Title,
			story.Description,
			story.AcceptanceCriteria,
			p.TestSpec,
			p.Context,
			e.cfg.PRDFile,
			p.CompletedCount(),
			len(p.Stories),
			story.DependsOn,
		)

		e.runner.Run(ctx, storyPrompt, outputCh)
		close(outputCh)

		updatedPRD, loadErr := e.store.Load(e.cfg)
		if loadErr != nil {
			logger.Error("failed to reload PRD after story, cannot continue", "error", loadErr, "story_id", story.ID)
			wrappedErr := fmt.Errorf("failed to reload PRD %s after story %s: %w", e.cfg.PRDFile, story.ID, loadErr)
			e.emit(EventError{Err: wrappedErr})
			return wrappedErr
		}

		updatedStory := updatedPRD.GetStory(story.ID)
		if updatedStory != nil && updatedStory.Passes {
			testsPass, _, _ := e.runTests(updatedPRD)
			if testsPass {
				logger.Debug("story completed", "story_id", story.ID)
				e.emit(EventStoryCompleted{Story: updatedStory, Success: true})
			} else {
				logger.Debug("tests failed despite passes flag, will retry", "story_id", story.ID)
				updatedStory.Passes = false
				if saveErr := e.store.Save(e.cfg, updatedPRD); saveErr != nil {
					logger.Warn("failed to save after test failure", "error", saveErr, "story_id", story.ID)
				}
				e.emit(EventStoryCompleted{Story: updatedStory, Success: false})
			}
		} else {
			e.emit(EventStoryCompleted{Story: story, Success: false})
		}
	}
}
