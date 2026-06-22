package prd

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

const (
	MaxContextSize   = 1 * 1024 * 1024 // 1MB max context to prevent memory exhaustion
	MaxStories       = 1000            // Maximum number of stories to prevent resource issues
	MaxStoryDescSize = 100 * 1024      // 100KB max story description
)

type Slice struct {
	ID           string `json:"id"`
	Behavior     string `json:"behavior"`
	RedHint      string `json:"red_hint"`
	RefactorHint string `json:"refactor_hint,omitempty"`
	Passes       bool   `json:"passes"`
}

type Story struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Slices      []*Slice `json:"slices,omitempty"`
	Priority    int      `json:"priority"`
	DependsOn   []string `json:"depends_on,omitempty"` // Story IDs this story depends on
	Passes      bool     `json:"passes"`
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

func (p *PRD) NextReadyStory() *Story {
	ready := p.ReadyStories()
	if len(ready) == 0 {
		return nil
	}
	sort.Slice(ready, func(i, j int) bool {
		if ready[i].Priority != ready[j].Priority {
			return ready[i].Priority < ready[j].Priority
		}
		return ready[i].ID < ready[j].ID
	})
	return ready[0]
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

func (s *Story) CompletedSliceCount() int {
	completed, _, _ := s.sliceProgress()
	return completed
}

func (s *Story) AllSlicesPassed() bool {
	_, _, allPassed := s.sliceProgress()
	return allPassed
}

func (s *Story) NextPendingSlice() *Slice {
	_, next, _ := s.sliceProgress()
	return next
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
	} else if len(p.Stories) > MaxStories {
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

func errLegacyAcceptanceCriteria(storyID string) error {
	if storyID == "" {
		storyID = "<unknown>"
	}
	return fmt.Errorf("story %q uses legacy acceptance_criteria; use slices instead", storyID)
}

func rejectLegacyAcceptanceCriteriaInJSON(data []byte) error {
	var raw struct {
		Stories []json.RawMessage `json:"stories"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}
	for _, storyData := range raw.Stories {
		var story struct {
			ID                 string          `json:"id"`
			AcceptanceCriteria json.RawMessage `json:"acceptance_criteria"`
		}
		if err := json.Unmarshal(storyData, &story); err != nil {
			continue
		}
		if len(story.AcceptanceCriteria) == 0 || string(story.AcceptanceCriteria) == "null" {
			continue
		}
		return errLegacyAcceptanceCriteria(story.ID)
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
		return errors.New("story title cannot be empty")
	}
	if len(s.Description) > MaxStoryDescSize {
		return fmt.Errorf("story description size %d exceeds maximum %d bytes", len(s.Description), MaxStoryDescSize)
	}
	if s.Priority < 0 {
		return fmt.Errorf("story priority %d cannot be negative", s.Priority)
	}
	if len(s.Slices) == 0 {
		return fmt.Errorf("story %q must have at least one slice", s.ID)
	}
	sliceIDs := make(map[string]bool)
	for i, sl := range s.Slices {
		if sl == nil {
			return fmt.Errorf("story %q slice %d cannot be nil", s.ID, i)
		}
		if sl.ID == "" {
			return fmt.Errorf("story %q slice %d id cannot be empty", s.ID, i)
		}
		if sliceIDs[sl.ID] {
			return fmt.Errorf("story %q has duplicate slice ID %q", s.ID, sl.ID)
		}
		sliceIDs[sl.ID] = true
		if sl.Behavior == "" {
			return fmt.Errorf("story %q slice %q behavior cannot be empty", s.ID, sl.ID)
		}
		if sl.RedHint == "" {
			return fmt.Errorf("story %q slice %q red hint cannot be empty", s.ID, sl.ID)
		}
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

func (s *Story) ResetSlicePasses() {
	for _, sl := range s.Slices {
		sl.Passes = false
	}
}

func (s *Story) sliceProgress() (completed int, next *Slice, allPassed bool) {
	allPassed = true
	for _, sl := range s.Slices {
		if sl.Passes {
			if next == nil {
				completed++
			}
			continue
		}
		if next == nil {
			next = sl
		}
		allPassed = false
	}
	return completed, next, allPassed
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
