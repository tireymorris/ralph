package prd

// Story represents a single user story in the PRD
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

// PRD represents the Product Requirements Document
type PRD struct {
	ProjectName string   `json:"project_name"`
	BranchName  string   `json:"branch_name,omitempty"`
	Stories     []*Story `json:"stories"`
}

// NextPendingStory returns the next story to implement
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

// CompletedCount returns the number of completed stories
func (p *PRD) CompletedCount() int {
	count := 0
	for _, story := range p.Stories {
		if story.Passes {
			count++
		}
	}
	return count
}

// FailedStories returns stories that have exceeded retry limits
func (p *PRD) FailedStories(maxRetries int) []*Story {
	var failed []*Story
	for _, story := range p.Stories {
		if !story.Passes && story.RetryCount >= maxRetries {
			failed = append(failed, story)
		}
	}
	return failed
}

// AllCompleted returns true if all stories are done
func (p *PRD) AllCompleted() bool {
	for _, story := range p.Stories {
		if !story.Passes {
			return false
		}
	}
	return true
}
