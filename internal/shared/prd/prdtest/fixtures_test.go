package prdtest

import "testing"

func TestFixtures(t *testing.T) {
	tests := []struct {
		name     string
		behavior string
	}{
		{name: "AC", behavior: "AC"},
		{name: "works", behavior: "works"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("Slices", func(t *testing.T) {
				slices := Slices(tt.behavior)
				if len(slices) != 1 {
					t.Fatalf("len(Slices) = %d, want 1", len(slices))
				}
				sl := slices[0]
				if sl.ID != "slice-1" {
					t.Errorf("slice ID = %q, want slice-1", sl.ID)
				}
				if sl.RedHint != "add failing test" {
					t.Errorf("slice RedHint = %q, want add failing test", sl.RedHint)
				}
				if sl.Behavior != tt.behavior {
					t.Errorf("slice Behavior = %q, want %q", sl.Behavior, tt.behavior)
				}
			})

			t.Run("StoryWithSlices", func(t *testing.T) {
				story := StoryWithSlices(tt.behavior)
				if story.Priority != 1 {
					t.Errorf("story Priority = %d, want 1", story.Priority)
				}
			})

			t.Run("SingleStoryPRD", func(t *testing.T) {
				p := SingleStoryPRD(tt.behavior)
				if p.ProjectName != "Test" {
					t.Errorf("ProjectName = %q, want Test", p.ProjectName)
				}
				if len(p.Stories) != 1 {
					t.Fatalf("Stories len = %d, want 1", len(p.Stories))
				}
				story := p.Stories[0]
				if story.ID != "1" || story.Title != "Story" || story.Description != "Desc" {
					t.Errorf("story = %+v, want id 1 title Story description Desc", story)
				}
				if story.Priority != 1 {
					t.Errorf("story Priority = %d, want 1", story.Priority)
				}
				if len(story.Slices) != 1 || story.Slices[0].Behavior != tt.behavior {
					t.Errorf("story Slices = %+v, want one slice with behavior %q", story.Slices, tt.behavior)
				}
			})
		})
	}
}
