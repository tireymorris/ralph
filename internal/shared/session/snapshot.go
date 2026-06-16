package session

import (
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runstate"
)

type RunSnapshot struct {
	Prompt           string
	Phase            string
	CurrentPRD       *prd.PRD
	CurrentStory     *prd.Story
	NextPendingSlice *prd.Slice
}

func (s *Session) RunSnapshot(fallbackPhase string) RunSnapshot {
	if s == nil || s.Driver == nil {
		return RunSnapshot{}
	}

	currentPRD := s.CurrentPRD()
	checkpoint := s.Checkpoint()
	phase := runstate.CheckpointPhase(checkpoint, currentPRD)
	if phase == runstate.PhaseReview && currentPRD == nil && checkpoint == "" && fallbackPhase != "" {
		phase = fallbackPhase
	}

	var currentStory *prd.Story
	var nextPendingSlice *prd.Slice
	if currentPRD != nil {
		currentStory = currentPRD.NextReadyStory()
		if currentStory != nil {
			nextPendingSlice = currentStory.NextPendingSlice()
		}
	}

	return RunSnapshot{
		Prompt:           s.Prompt(),
		Phase:            phase,
		CurrentPRD:       currentPRD,
		CurrentStory:     currentStory,
		NextPendingSlice: nextPendingSlice,
	}
}
