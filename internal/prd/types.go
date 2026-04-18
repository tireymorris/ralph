package prd

import (
	"encoding/json"
	"fmt"
)

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
	Priority           int      `json:"priority"`
	DependsOn          []string `json:"depends_on,omitempty"` // Story IDs this story depends on
	Passes             bool     `json:"passes"`
	RetryCount         int      `json:"retry_count"`
}

type PRD struct {
	Version     int64    `json:"version"` // Incremented on each save for optimistic locking
	ProjectName string   `json:"project_name"`
	BranchName  string   `json:"branch_name,omitempty"`
	Context     string   `json:"context,omitempty"`
	TestSpec    string   `json:"test_spec,omitempty"` // Holistic test spec covering all stories
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

// ReadyStories returns all stories that are ready to run (not passed, not exceeded
// retries, and all dependencies are satisfied).
func (p *PRD) ReadyStories(maxRetries int) []*Story {
	var ready []*Story
	for _, story := range p.Stories {
		if story.Passes {
			continue
		}
		if story.RetryCount >= maxRetries {
			continue
		}
		if !p.dependenciesSatisfied(story) {
			continue
		}
		ready = append(ready, story)
	}
	return ready
}

// dependenciesSatisfied returns true if all dependencies for the story are complete.
func (p *PRD) dependenciesSatisfied(story *Story) bool {
	depMap := make(map[string]bool)
	for _, s := range p.Stories {
		depMap[s.ID] = s.Passes
	}
	for _, depID := range story.DependsOn {
		if passed, ok := depMap[depID]; !ok || !passed {
			return false
		}
	}
	return true
}

// BlockedStories returns stories that cannot run due to failed dependencies.
func (p *PRD) BlockedStories(maxRetries int) []*Story {
	var blocked []*Story
	for _, story := range p.Stories {
		if story.Passes {
			continue
		}
		if story.RetryCount >= maxRetries {
			continue
		}
		if !p.dependenciesSatisfied(story) {
			blocked = append(blocked, story)
		}
	}
	return blocked
}

// ValidateDependencies checks for circular dependencies in the story graph.
func (p *PRD) ValidateDependencies() error {
	visited := make(map[string]bool)
	var dfs func(id string, path []string) error
	dfs = func(id string, path []string) error {
		if id == "" {
			return nil
		}
		for _, p := range path {
			if p == id {
				return fmt.Errorf("circular dependency detected: %s", append(path, id))
			}
		}
		if visited[id] {
			return nil
		}
		visited[id] = true
		story := p.GetStory(id)
		if story == nil {
			return fmt.Errorf("story %q depends on non-existent story", id)
		}
		for _, depID := range story.DependsOn {
			if err := dfs(depID, append(path, id)); err != nil {
				return err
			}
		}
		return nil
	}

	for _, story := range p.Stories {
		if err := dfs(story.ID, nil); err != nil {
			return err
		}
	}
	return nil
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

	if err := p.ValidateDependencies(); err != nil {
		return fmt.Errorf("invalid dependencies: %w", err)
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

	for _, dep := range s.DependsOn {
		if dep == "" {
			return fmt.Errorf("story %q has empty dependency ID", s.ID)
		}
		if dep == s.ID {
			return fmt.Errorf("story %q cannot depend on itself", s.ID)
		}
	}

	return nil
}

func (p *PRD) ToJSON() string {
	data, _ := json.MarshalIndent(p, "", "  ")
	return string(data)
}
