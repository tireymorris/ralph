package prd

import "testing"

func TestNextReadyStoryRespectsDependencies(t *testing.T) {
	p := &PRD{Stories: []*Story{
		{ID: "story-2", Priority: 1, Passes: false, DependsOn: []string{"story-1"}},
		{ID: "story-1", Priority: 2, Passes: false},
	}}

	got := p.NextReadyStory()
	if got == nil {
		t.Fatal("NextReadyStory() = nil, want story-1")
	}
	if got.ID != "story-1" {
		t.Fatalf("NextReadyStory().ID = %q, want story-1", got.ID)
	}
}

func TestNextReadyStorySortsByPriorityThenID(t *testing.T) {
	p := &PRD{Stories: []*Story{
		{ID: "story-b", Priority: 1, Passes: false},
		{ID: "story-a", Priority: 1, Passes: false},
		{ID: "story-c", Priority: 2, Passes: false},
	}}

	got := p.NextReadyStory()
	if got == nil {
		t.Fatal("NextReadyStory() = nil")
	}
	if got.ID != "story-a" {
		t.Fatalf("NextReadyStory().ID = %q, want story-a (priority 1, then ID)", got.ID)
	}
}

func TestNextReadyStoryReturnsNilWhenBlocked(t *testing.T) {
	p := &PRD{Stories: []*Story{
		{ID: "story-2", Priority: 1, Passes: false, DependsOn: []string{"story-1"}},
	}}

	if got := p.NextReadyStory(); got != nil {
		t.Fatalf("NextReadyStory() = %+v, want nil when all incomplete stories are blocked", got)
	}
}
