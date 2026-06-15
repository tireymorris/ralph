package prd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func (p *PRD) Snapshot(path string) error {
	if err := p.Validate(); err != nil {
		return fmt.Errorf("PRD validation failed before snapshot %q: %w", path, err)
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal PRD snapshot %q: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create snapshot directory %q: %w", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write PRD snapshot %q: %w", path, err)
	}

	return nil
}

func (p *PRD) UnmarkAllStories() {
	for _, story := range p.Stories {
		story.Passes = false
		story.ResetSlicePasses()
	}
}

func (p *PRD) UnmarkStories(ids []string) error {
	for _, id := range ids {
		story := p.GetStory(id)
		if story == nil {
			return fmt.Errorf("unknown story id %q", id)
		}
		story.Passes = false
	}
	return nil
}
