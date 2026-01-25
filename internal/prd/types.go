package prd

type Story struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	TestSpec           string   `json:"test_spec,omitempty"`
	Priority           int      `json:"priority"`
	Passes             bool     `json:"passes"`
	RetryCount         int      `json:"retry_count"`
}

type PRD struct {
	Version     int64    `json:"version"` // Incremented on each save for optimistic locking
	ProjectName string   `json:"project_name"`
	BranchName  string   `json:"branch_name,omitempty"`
	Context     string   `json:"context,omitempty"`
	Stories     []*Story `json:"stories"`
}

func (p *PRD) NextPendingStory(maxRetries int) *Story {
	var best *Story
	for _, story := range p.Stories {
		if story.Passes {
			continue
		}
		if story.RetryCount >= maxRetries {
			continue
		}
		if best == nil || story.Priority < best.Priority {
			best = story
		}
	}
	return best
}

func (p *PRD) CompletedCount() int {
	count := 0
	for _, story := range p.Stories {
		if story.Passes {
			count++
		}
	}
	return count
}

func (p *PRD) FailedStories(maxRetries int) []*Story {
	var failed []*Story
	for _, story := range p.Stories {
		if !story.Passes && story.RetryCount >= maxRetries {
			failed = append(failed, story)
		}
	}
	return failed
}

func (p *PRD) AllCompleted() bool {
	for _, story := range p.Stories {
		if !story.Passes {
			return false
		}
	}
	return true
}

func (p *PRD) GetStory(id string) *Story {
	for _, story := range p.Stories {
		if story.ID == id {
			return story
		}
	}
	return nil
}
