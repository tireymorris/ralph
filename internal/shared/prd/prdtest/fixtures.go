package prdtest

import "ralph/internal/shared/prd"

func Slices(behavior string) []*prd.Slice {
	return []*prd.Slice{{
		ID:       "slice-1",
		Behavior: behavior,
		RedHint:  "add failing test",
	}}
}

func StoryWithSlices(behavior string) *prd.Story {
	return &prd.Story{
		ID:          "1",
		Title:       "Story",
		Description: "Desc",
		Slices:      Slices(behavior),
		Priority:    1,
	}
}

func SingleStoryPRD(behavior string) *prd.PRD {
	return &prd.PRD{
		ProjectName: "Test",
		Stories:     []*prd.Story{StoryWithSlices(behavior)},
	}
}
