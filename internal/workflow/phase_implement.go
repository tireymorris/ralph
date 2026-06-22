package workflow

import (
	"context"
	"fmt"
	"strings"

	"ralph/internal/shared/logger"
	"ralph/internal/shared/prd"
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

			e.resetRecoveryAttempts()
			if err := e.runTestGateWithRecovery(ctx, p); err != nil {
				return err
			}

			e.emit(EventCompleted{})
			return nil
		}

		story := p.NextReadyStory()
		if story == nil {
			blocked := p.BlockedStories()
			if len(blocked) > 0 {
				logger.Error("no ready stories, all incomplete stories are dependency-blocked", "blocked_count", len(blocked))
				blockedErr := fmt.Errorf("all incomplete stories are dependency-blocked: %s", describeBlockedStories(p, blocked))
				e.emit(EventError{Err: blockedErr})
				return blockedErr
			}
			continue
		}

		logger.Debug("starting story",
			"story_id", story.ID,
			"title", story.Title)

		e.emit(EventStoryStarted{Story: story})

		updatedPRD, updatedStory, sliceErr := e.runStorySlices(ctx, p, story)
		if sliceErr != nil {
			logger.Error("implementation runner failed", "error", sliceErr, "story_id", story.ID)
			e.emit(EventError{Err: sliceErr})
			return sliceErr
		}

		logger.Debug("story completed", "story_id", story.ID)
		e.emit(EventStoryCompleted{Story: updatedStory, Success: true})

		e.resetRecoveryAttempts()
		blocked, reviewErr := e.runImplementationReview(ctx, updatedPRD)
		if reviewErr != nil {
			return reviewErr
		}
		if blocked {
			return nil
		}

		e.resetRecoveryAttempts()
		if err := e.runTestGateWithRecovery(ctx, updatedPRD); err != nil {
			return err
		}
	}
}
