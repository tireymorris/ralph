package prd

import (
	"errors"
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
}

type PRD struct {
	Version     int64    `json:"version"` // Incremented on each save for optimistic locking
	ProjectName string   `json:"project_name"`
	BranchName  string   `json:"branch_name,omitempty"`
	Context     string   `json:"context,omitempty"`
	TestSpec    string   `json:"test_spec,omitempty"`    // Holistic test spec covering all stories
	TestCommand string   `json:"test_command,omitempty"` // Project-specific test command (overrides config)
	Stories     []*Story `json:"stories"`
}

func (p *PRD) NextPendingStory() *Story {
	var best *Story
	for _, story := range p.Stories {
		if story.Passes {
			continue
		}
		if best == nil || story.Priority < best.Priority {
			best = story
		}
	}
	return best
}

func (p *PRD) ReadyStories() []*Story {
	var ready []*Story
	for _, story := range p.Stories {
		if story.Passes || !p.dependenciesSatisfied(story) {
			continue
		}
		ready = append(ready, story)
	}
	return ready
}

func (p *PRD) BlockedStories() []*Story {
	var blocked []*Story
	for _, story := range p.Stories {
		if story.Passes || p.dependenciesSatisfied(story) {
			continue
		}
		blocked = append(blocked, story)
	}
	return blocked
}

func (p *PRD) ValidateDependencies() error {
	visited := make(map[string]bool)
	var dfs func(id string, path []string) error
	dfs = func(id string, path []string) error {
		if id == "" {
			return nil
		}
		for _, pathID := range path {
			if pathID == id {
				cycle := append(path, id)
				return fmt.Errorf("circular dependency detected: %v", cycle)
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

func (s *Story) Validate(seenIDs map[string]bool) error {
	if s.ID == "" {
		return errors.New("story ID cannot be empty")
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

func (p *PRD) dependenciesSatisfied(story *Story) bool {
	depMap := p.storyPassMap()
	for _, depID := range story.DependsOn {
		if passed, ok := depMap[depID]; !ok || !passed {
			return false
		}
	}
	return true
}

func (p *PRD) storyPassMap() map[string]bool {
	depMap := make(map[string]bool, len(p.Stories))
	for _, s := range p.Stories {
		depMap[s.ID] = s.Passes
	}
	return depMap
}
