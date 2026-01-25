package prd

import "fmt"

const (
	MaxContextSize        = 1 * 1024 * 1024 // 1MB max context to prevent memory exhaustion
	MaxStories            = 1000            // Maximum number of stories to prevent resource issues
	MaxStoryDescSize      = 100 * 1024      // 100KB max story description
	MaxAcceptanceCriteria = 50              // Maximum acceptance criteria per story
)

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

// Validate validates the PRD data structure and content.
func (p *PRD) Validate() error {
	if len(p.Context) > MaxContextSize {
		return fmt.Errorf("context size %d exceeds maximum %d bytes", len(p.Context), MaxContextSize)
	}

	if len(p.Stories) > MaxStories {
		return fmt.Errorf("story count %d exceeds maximum %d", len(p.Stories), MaxStories)
	}

	seenIDs := make(map[string]bool)
	for i, story := range p.Stories {
		if err := story.Validate(seenIDs); err != nil {
			return fmt.Errorf("story %d (%q): %w", i, story.ID, err)
		}
		seenIDs[story.ID] = true
	}

	return nil
}

// Validate validates a single story and checks for duplicate IDs.
func (s *Story) Validate(seenIDs map[string]bool) error {
	if s.ID == "" {
		return fmt.Errorf("story ID cannot be empty")
	}

	if seenIDs[s.ID] {
		return fmt.Errorf("duplicate story ID %q", s.ID)
	}

	if s.Title == "" {
		return fmt.Errorf("story title cannot be empty")
	}

	if len(s.Description) > MaxStoryDescSize {
		return fmt.Errorf("story description size %d exceeds maximum %d bytes", len(s.Description), MaxStoryDescSize)
	}

	if s.Priority < 0 {
		return fmt.Errorf("story priority %d cannot be negative", s.Priority)
	}

	if len(s.AcceptanceCriteria) > MaxAcceptanceCriteria {
		return fmt.Errorf("story has %d acceptance criteria, maximum %d", len(s.AcceptanceCriteria), MaxAcceptanceCriteria)
	}

	if s.RetryCount < 0 {
		return fmt.Errorf("story retry count %d cannot be negative", s.RetryCount)
	}

	return nil
}
