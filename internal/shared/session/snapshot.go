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
	CompletedStories int
	TotalStories     int
	Activity         RunActivity
}

type RunActivityKind string

const (
	ActivityImplementing RunActivityKind = "implementing"
	ActivityReview       RunActivityKind = "review"
	ActivityRecovery     RunActivityKind = "recovery"
	ActivityCleanup      RunActivityKind = "cleanup"
)

type RunActivity struct {
	Kind         RunActivityKind
	StoryID      string
	StoryTitle   string
	Iteration    int
	Attempt      int
	MaxAttempts  int
	FindingCount int
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
	completedStories := 0
	totalStories := 0
	if currentPRD != nil {
		if progress := currentPRD.RunProgress(); progress != nil {
			completedStories = progress.Completed
			totalStories = progress.Total
		}
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
		CompletedStories: completedStories,
		TotalStories:     totalStories,
	}
}
