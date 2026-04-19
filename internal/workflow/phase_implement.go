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

	iteration := 0

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

		ready := p.ReadyStories(e.cfg.RetryAttempts)
		if len(ready) == 0 {
			failed := p.FailedStories(e.cfg.RetryAttempts)
			if len(failed) > 0 {
				logger.Error("all remaining stories have failed", "failed_count", len(failed))
				e.emit(EventFailed{FailedStories: failed})
				return fmt.Errorf("all remaining stories have failed (%d stories)", len(failed))
			}
			blocked := p.BlockedStories(e.cfg.RetryAttempts)
			if len(blocked) > 0 {
				logger.Error("stories blocked by dependencies", "blocked_count", len(blocked))
				e.emit(EventFailed{FailedStories: blocked})
				return fmt.Errorf("%d stories blocked by dependencies and cannot be completed", len(blocked))
			}
		}

		story := ready[0]

		iteration++
		if iteration > e.cfg.MaxIterations {
			logger.Error("max iterations reached", "iterations", iteration, "max", e.cfg.MaxIterations)
			e.emit(EventFailed{FailedStories: p.FailedStories(e.cfg.RetryAttempts)})
			return fmt.Errorf("max iterations (%d) reached after %d iterations", e.cfg.MaxIterations, iteration)
		}

		logger.Debug("starting story",
			"story_id", story.ID,
			"title", story.Title,
			"iteration", iteration,
			"retry_count", story.RetryCount)

		gitState, saveErr := e.saveGitState()
		if saveErr != nil {
			logger.Warn("failed to save git state", "error", saveErr, "story_id", story.ID)
		}

		e.emit(EventStoryStarted{Story: story, Iteration: iteration})

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
			iteration,
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
			testsPass, testOutput, _ := e.runTests(updatedPRD)
			if testsPass {
				logger.Debug("story marked as completed", "story_id", story.ID)
				e.emit(EventStoryCompleted{Story: updatedStory, Success: true})
			} else {
				logger.Debug("tests failed despite passes flag", "story_id", story.ID, "test_output", testOutput)
				if gitState != "" {
					if rollbackErr := e.rollbackToState(gitState); rollbackErr != nil {
						logger.Error("rollback failed", "error", rollbackErr, "story_id", story.ID)
						e.emit(EventError{Err: fmt.Errorf("story %s test verification failed and rollback failed: %w", story.ID, rollbackErr)})
						return fmt.Errorf("story %s test verification failed and rollback failed: %w", story.ID, rollbackErr)
					}
					logger.Info("rolled back after test verification failure", "story_id", story.ID, "git_state", gitState)
				}
				updatedStory.Passes = false
				updatedStory.RetryCount++
				if saveErr := e.store.Save(e.cfg, updatedPRD); saveErr != nil {
					logger.Warn("failed to save after test failure", "error", saveErr, "story_id", story.ID)
				}
				e.emit(EventStoryCompleted{Story: updatedStory, Success: false})
			}
		} else {
			testsPass, testOutput, _ := e.runTests(updatedPRD)
			if testsPass {
				logger.Debug("tests pass but passes false - AI under-reported", "story_id", story.ID)
				if gitState != "" {
					if rollbackErr := e.rollbackToState(gitState); rollbackErr != nil {
						logger.Error("rollback failed", "error", rollbackErr, "story_id", story.ID)
						e.emit(EventError{Err: fmt.Errorf("story %s retry failed: %w", story.ID, rollbackErr)})
						return fmt.Errorf("story %s retry failed: %w", story.ID, rollbackErr)
					}
					logger.Info("rolled back for retry after test verification", "story_id", story.ID, "git_state", gitState)
				}
				if updatedStory != nil {
					updatedStory.RetryCount++
					if saveErr := e.store.Save(e.cfg, updatedPRD); saveErr != nil {
						logger.Warn("failed to save retry count", "error", saveErr, "story_id", story.ID)
					}
				}
				e.emit(EventStoryCompleted{Story: updatedStory, Success: false})
			} else {
				logger.Debug("story not completed, tests failed", "story_id", story.ID, "test_output", testOutput)
				if gitState != "" {
					if rollbackErr := e.rollbackToState(gitState); rollbackErr != nil {
						logger.Error("rollback failed", "error", rollbackErr, "story_id", story.ID)
						e.emit(EventError{Err: fmt.Errorf("story %s failed and rollback failed: %w", story.ID, rollbackErr)})
						return fmt.Errorf("story %s failed and rollback failed: %w", story.ID, rollbackErr)
					} else {
						logger.Info("rolled back after story failure", "story_id", story.ID, "git_state", gitState)
					}
				}

				if updatedStory != nil {
					updatedStory.RetryCount++
					if saveErr := e.store.Save(e.cfg, updatedPRD); saveErr != nil {
						logger.Warn("failed to save retry count", "error", saveErr, "story_id", story.ID)
					}
				}
				e.emit(EventStoryCompleted{Story: story, Success: false})
			}
		}
	}
}
