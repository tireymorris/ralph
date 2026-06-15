package prd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnmarkStories_knownID(t *testing.T) {
	p := &PRD{Stories: []*Story{{ID: "story-1", Passes: true}}}
	if err := p.UnmarkStories([]string{"story-1"}); err != nil {
		t.Fatalf("UnmarkStories: %v", err)
	}
	if p.GetStory("story-1").Passes {
		t.Error("expected story-1 Passes false")
	}
}

func TestUnmarkStories_unknownID(t *testing.T) {
	p := &PRD{Stories: []*Story{{ID: "story-1", Passes: true}}}
	err := p.UnmarkStories([]string{"missing-story"})
	if err == nil {
		t.Fatal("expected error for unknown story id")
	}
	if !strings.Contains(err.Error(), "missing-story") {
		t.Errorf("error %q should contain story id", err)
	}
}

func TestUnmarkAllStories(t *testing.T) {
	p := &PRD{Stories: []*Story{
		{ID: "story-1", Passes: true, Slices: []*Slice{{ID: "slice-1", Behavior: "one", RedHint: "red", Passes: true}}},
		{ID: "story-2", Passes: true, Slices: []*Slice{{ID: "slice-2", Behavior: "two", RedHint: "red", Passes: true}}},
	}}
	p.UnmarkAllStories()
	for _, id := range []string{"story-1", "story-2"} {
		if p.GetStory(id).Passes {
			t.Errorf("expected %s Passes false", id)
		}
		if got := p.GetStory(id).CompletedSliceCount(); got != 0 {
			t.Errorf("expected %s slice passes reset, got %d", id, got)
		}
	}
}

func TestSnapshot_writesLoadableJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".ralph", "runs", "run-1", "prd.snapshot.json")

	original := &PRD{
		ProjectName: "snap project",
		Version:     3,
		Stories: []*Story{
			{ID: "story-1", Title: "One", Priority: 1, Passes: true},
		},
	}
	if err := original.Validate(); err != nil {
		t.Fatalf("Validate original: %v", err)
	}

	if err := original.Snapshot(path); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}

	var loaded PRD
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	if err := loaded.Validate(); err != nil {
		t.Fatalf("Validate loaded: %v", err)
	}
	if loaded.ProjectName != original.ProjectName {
		t.Errorf("ProjectName = %q, want %q", loaded.ProjectName, original.ProjectName)
	}
	if loaded.Version != original.Version {
		t.Errorf("Version = %d, want %d", loaded.Version, original.Version)
	}
	if len(loaded.Stories) != 1 || loaded.Stories[0].ID != "story-1" || !loaded.Stories[0].Passes {
		t.Errorf("loaded story: %+v", loaded.Stories)
	}
}
