package workflow

import (
	"context"
	"fmt"

	"ralph/internal/prompt"
	"ralph/internal/shared/gitdiff"
	"ralph/internal/shared/prd"
	"ralph/internal/workflow/events"
)

func storyImplementationSlices(story *prd.Story) []prompt.SliceData {
	slices := story.Slices
	if len(slices) == 0 && len(story.AcceptanceCriteria) > 0 {
		slices = make([]*prd.Slice, 0, len(story.AcceptanceCriteria))
		for i, criterion := range story.AcceptanceCriteria {
			slices = append(slices, &prd.Slice{
				ID:       fmt.Sprintf("slice-%d", i+1),
				Behavior: criterion,
				RedHint:  fmt.Sprintf("add failing test for: %s", criterion),
			})
		}
	}
	result := make([]prompt.SliceData, 0, len(slices))
	for _, slice := range slices {
		if slice == nil {
			continue
		}
		result = append(result, prompt.SliceData{
			ID:           slice.ID,
			Behavior:     slice.Behavior,
			RedHint:      slice.RedHint,
			RefactorHint: slice.RefactorHint,
			Passes:       slice.Passes,
		})
	}
	return result
}

func storySliceByID(story *prd.Story, id string) *prd.Slice {
	for _, slice := range story.Slices {
		if slice != nil && slice.ID == id {
			return slice
		}
	}
	return nil
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
			storyImplementationSlices(story),
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

		committed, commitErr := gitdiff.CommitChangedFiles(e.cfg.WorkDir, fmt.Sprintf("ralph: %s", story.ID))
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
		updatedSlice := storySliceByID(updatedStory, currentSlice.ID)
		if updatedSlice == nil {
			return nil, nil, fmt.Errorf("story %s missing slice %s after implementation", story.ID, currentSlice.ID)
		}
		if !updatedSlice.Passes {
			updatedSlice.Passes = true
		}

		updatedStory.Passes = updatedStory.AllSlicesPassed()
		if saveErr := e.store.Save(e.cfg, updatedPRD); saveErr != nil {
			return nil, nil, fmt.Errorf("failed to save PRD after completing story %s slice %s: %w", story.ID, currentSlice.ID, saveErr)
		}

		p = updatedPRD
		story = updatedStory
	}
}
