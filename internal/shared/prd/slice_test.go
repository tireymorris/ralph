package prd

import "testing"

func TestStoryNextPendingSlice(t *testing.T) {
	story := &Story{
		ID: "story-1",
		Slices: []*Slice{
			{ID: "slice-1", Behavior: "first", RedHint: "write failing test", Passes: true},
			{ID: "slice-2", Behavior: "second", RedHint: "write failing test", Passes: false},
		},
	}

	got := story.NextPendingSlice()
	if got == nil {
		t.Fatal("NextPendingSlice() = nil, want pending slice")
	}
	if got.ID != "slice-2" {
		t.Fatalf("NextPendingSlice().ID = %q, want %q", got.ID, "slice-2")
	}
}

func TestStorySliceProgress(t *testing.T) {
	story := &Story{
		ID: "story-1",
		Slices: []*Slice{
			{ID: "slice-1", Behavior: "first", RedHint: "write failing test", Passes: true},
			{ID: "slice-2", Behavior: "second", RedHint: "write failing test", Passes: false},
		},
	}

	if got := story.CompletedSliceCount(); got != 1 {
		t.Fatalf("CompletedSliceCount() = %d, want 1", got)
	}
	if story.AllSlicesPassed() {
		t.Fatal("AllSlicesPassed() = true, want false")
	}

	story.Slices[1].Passes = true

	if got := story.CompletedSliceCount(); got != 2 {
		t.Fatalf("CompletedSliceCount() = %d, want 2", got)
	}
	if !story.AllSlicesPassed() {
		t.Fatal("AllSlicesPassed() = false, want true")
	}
}
