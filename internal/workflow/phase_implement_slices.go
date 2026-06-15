package workflow

import (
	"context"
	"fmt"

	"ralph/internal/prompt"
	"ralph/internal/shared/gitdiff"
	"ralph/internal/shared/prd"
	"ralph/internal/workflow/events"
)

var commitChangedFiles = gitdiff.CommitChangedFiles

func storyImplementationSliceData(slice *prd.Slice) []prompt.SliceData {
	if slice == nil {
		return nil
	}
	return []prompt.SliceData{{
		ID:           slice.ID,
		Behavior:     slice.Behavior,
		RedHint:      slice.RedHint,
		RefactorHint: slice.RefactorHint,
		Passes:       slice.Passes,
	}}
}

func (e *Executor) runStorySlices(ctx context.Context, p *prd.PRD, story *prd.Story) (*prd.PRD, *prd.Story, error) {
	for {
		currentSlice := story.NextPendingSlice()
		if currentSlice == nil {
			if !story.Passes {
				story.Passes = story.AllSlicesPassed()
				if err := e.store.Save(e.cfg, p); err != nil {
					return nil, nil, fmt.Errorf("failed to save PRD after completing story %s: %w", story.ID, err)
				}
			}
			return p, story, nil
		}

		storyPrompt := prompt.StoryImplementation(
			story.ID,
			story.Title,
			story.Description,
			storyImplementationSliceData(currentSlice),
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
				return nil, nil, recErr
			}
			if recovered {
				runErr = e.runWithForwardedOutput(ctx, storyPrompt)
			}
			if runErr != nil {
				return nil, nil, fmt.Errorf("implementation failed for story %s slice %s: %w", story.ID, currentSlice.ID, runErr)
			}
		}

		committed, commitErr := commitChangedFiles(e.cfg.WorkDir, fmt.Sprintf("ralph: %s/%s", story.ID, currentSlice.ID))
		if commitErr != nil {
			return nil, nil, fmt.Errorf("commit story %s slice %s changes: %w", story.ID, currentSlice.ID, commitErr)
		}
		if committed {
			e.emit(EventOutput{Output: events.Output{Text: fmt.Sprintf("Committed story %s slice %s changes before next slice.", story.ID, currentSlice.ID)}})
		}

		updatedPRD, loadErr := e.store.Load(e.cfg)
		if loadErr != nil {
			return nil, nil, fmt.Errorf("failed to reload PRD %s after story %s slice %s: %w", e.cfg.PRDFile, story.ID, currentSlice.ID, loadErr)
		}
		updatedStory := updatedPRD.GetStory(story.ID)
		if updatedStory == nil {
			return nil, nil, fmt.Errorf("story %s disappeared after slice %s implementation", story.ID, currentSlice.ID)
		}
		var updatedSlice *prd.Slice
		for _, slice := range updatedStory.Slices {
			if slice != nil && slice.ID == currentSlice.ID {
				updatedSlice = slice
				break
			}
		}
		if updatedSlice == nil {
			return nil, nil, fmt.Errorf("story %s missing slice %s after implementation", story.ID, currentSlice.ID)
		}
		if !updatedSlice.Passes {
			updatedSlice.Passes = true
		}

		if saveErr := e.store.Save(e.cfg, updatedPRD); saveErr != nil {
			return nil, nil, fmt.Errorf("failed to save PRD after completing story %s slice %s: %w", story.ID, currentSlice.ID, saveErr)
		}

		p = updatedPRD
		story = updatedStory
	}
}
