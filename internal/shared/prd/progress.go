package prd

type RunProgress struct {
	Completed int                `json:"completed"`
	Total     int                `json:"total"`
	Stories   []RunProgressStory `json:"stories,omitempty"`
}

type RunProgressStory struct {
	ID              string             `json:"id"`
	Title           string             `json:"title"`
	Passes          bool               `json:"passes"`
	CompletedSlices int                `json:"completed_slices"`
	TotalSlices     int                `json:"total_slices"`
	Slices          []RunProgressSlice `json:"slices,omitempty"`
}

type RunProgressSlice struct {
	ID           string `json:"id"`
	Behavior     string `json:"behavior"`
	RedHint      string `json:"red_hint"`
	RefactorHint string `json:"refactor_hint,omitempty"`
	Passes       bool   `json:"passes"`
}

func (p *PRD) RunProgress() *RunProgress {
	if p == nil {
		return nil
	}

	progress := &RunProgress{
		Completed: p.CompletedCount(),
		Total:     len(p.Stories),
		Stories:   make([]RunProgressStory, 0, len(p.Stories)),
	}
	for _, story := range p.Stories {
		progress.Stories = append(progress.Stories, story.RunProgress())
	}
	return progress
}

func (s *Story) RunProgress() RunProgressStory {
	progress := RunProgressStory{
		ID:              s.ID,
		Title:           s.Title,
		Passes:          s.Passes,
		CompletedSlices: s.CompletedSliceCount(),
		TotalSlices:     len(s.Slices),
		Slices:          make([]RunProgressSlice, 0, len(s.Slices)),
	}
	for _, slice := range s.Slices {
		if slice == nil {
			continue
		}
		progress.Slices = append(progress.Slices, RunProgressSlice{
			ID:           slice.ID,
			Behavior:     slice.Behavior,
			RedHint:      slice.RedHint,
			RefactorHint: slice.RefactorHint,
			Passes:       slice.Passes,
		})
	}
	return progress
}
