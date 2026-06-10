package workflow

import (
	"context"
	"fmt"
	"strings"

	"ralph/internal/prompt"
	"ralph/internal/shared/gitdiff"
	"ralph/internal/shared/logger"
	"ralph/internal/shared/prd"
	"ralph/internal/workflow/events"
)

func describeBlockedStories(p *prd.PRD, blocked []*prd.Story) string {
	descriptions := make([]string, 0, len(blocked))
	for _, story := range blocked {
		var unsatisfied []string
		for _, depID := range story.DependsOn {
			dep := p.GetStory(depID)
			if dep == nil || !dep.Passes {
				unsatisfied = append(unsatisfied, depID)
			}
		}
		descriptions = append(descriptions, fmt.Sprintf("%s (depends on: %s)", story.ID, strings.Join(unsatisfied, ", ")))
	}
	return strings.Join(descriptions, "; ")
}

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
			if !e.cfg.SkipCleanup {
				if err := e.RunCleanup(ctx, p); err != nil {
					return err
				}
			}
			e.emit(EventCompleted{})
			return nil
		}

		ready := p.ReadyStories()
		if len(ready) == 0 {
			blocked := p.BlockedStories()
			if len(blocked) > 0 {
				logger.Error("no ready stories, all incomplete stories are dependency-blocked", "blocked_count", len(blocked))
				blockedErr := fmt.Errorf("all incomplete stories are dependency-blocked: %s", describeBlockedStories(p, blocked))
				e.emit(EventError{Err: blockedErr})
				return blockedErr
			}
			continue
		}

		story := ready[0]

		logger.Debug("starting story",
			"story_id", story.ID,
			"title", story.Title)

		e.emit(EventStoryStarted{Story: story})

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

		runErr := e.runWithForwardedOutput(ctx, storyPrompt)
		if runErr != nil {
			recovered, recErr := e.runRecovery(ctx, p, prompt.RecoveryReasonStoryFailure, runErr.Error(), nil)
			if recErr != nil {
				logger.Error("implementation recovery failed", "error", recErr, "story_id", story.ID)
				e.emit(EventError{Err: fmt.Errorf("implementation recovery failed for story %s: %w", story.ID, recErr)})
				return fmt.Errorf("implementation recovery failed for story %s: %w", story.ID, recErr)
			}
			if recovered {
				runErr = e.runWithForwardedOutput(ctx, storyPrompt)
			}
			if runErr != nil {
				logger.Error("implementation runner failed", "error", runErr, "story_id", story.ID)
				e.emit(EventError{Err: fmt.Errorf("implementation failed for story %s: %w", story.ID, runErr)})
				return fmt.Errorf("implementation failed for story %s: %w", story.ID, runErr)
			}
		}

		committed, commitErr := gitdiff.CommitChangedFiles(e.cfg.WorkDir, fmt.Sprintf("ralph: %s", story.ID))
		if commitErr != nil {
			logger.Error("failed to commit story changes", "error", commitErr, "story_id", story.ID)
			e.emit(EventError{Err: fmt.Errorf("commit story %s changes: %w", story.ID, commitErr)})
			return fmt.Errorf("commit story %s changes: %w", story.ID, commitErr)
		}
		if committed {
			e.emit(EventOutput{Output: events.Output{Text: fmt.Sprintf("Committed story %s changes before review.", story.ID)}})
		}

		updatedPRD, loadErr := e.store.Load(e.cfg)
		if loadErr != nil {
			logger.Error("failed to reload PRD after story, cannot continue", "error", loadErr, "story_id", story.ID)
			wrappedErr := fmt.Errorf("failed to reload PRD %s after story %s: %w", e.cfg.PRDFile, story.ID, loadErr)
			e.emit(EventError{Err: wrappedErr})
			return wrappedErr
		}

		updatedStory := updatedPRD.GetStory(story.ID)
		if updatedStory == nil {
			logger.Error("story disappeared after implementation", "story_id", story.ID)
			e.emit(EventStoryCompleted{Story: story, Success: false})
			continue
		}

		updatedStory.Passes = true
		if saveErr := e.store.Save(e.cfg, updatedPRD); saveErr != nil {
			logger.Warn("failed to save PRD after marking story complete", "error", saveErr, "story_id", story.ID)
			e.emit(EventError{Err: fmt.Errorf("failed to save PRD after completing story %s: %w", story.ID, saveErr)})
			return fmt.Errorf("failed to save PRD after completing story %s: %w", story.ID, saveErr)
		}

		logger.Debug("story completed", "story_id", story.ID)
		e.emit(EventStoryCompleted{Story: updatedStory, Success: true})

		if _, reviewErr := e.runImplementationReview(ctx, updatedPRD); reviewErr != nil {
			return reviewErr
		}
	}
}
