package prd

import "testing"

func TestRunProgress(t *testing.T) {
	p := &PRD{
		Stories: []*Story{
			{
				ID:     "s1",
				Title:  "done",
				Passes: true,
				Slices: []*Slice{
					{ID: "slice-1", Behavior: "first", RedHint: "test it", Passes: true},
				},
			},
			{
				ID:    "s2",
				Title: "todo",
				Slices: []*Slice{
					{ID: "slice-1", Behavior: "second", RedHint: "test it", RefactorHint: "extract helper"},
				},
			},
		},
	}

	progress := p.RunProgress()
	if progress == nil {
		t.Fatal("RunProgress() = nil")
	}
	if progress.Completed != 1 {
		t.Fatalf("RunProgress().Completed = %d, want 1", progress.Completed)
	}
	if progress.Total != 2 {
		t.Fatalf("RunProgress().Total = %d, want 2", progress.Total)
	}
	if len(progress.Stories) != 2 {
		t.Fatalf("RunProgress().Stories = %d, want 2", len(progress.Stories))
	}
	if progress.Stories[1].CompletedSlices != 0 {
		t.Fatalf("RunProgress().Stories[1].CompletedSlices = %d, want 0", progress.Stories[1].CompletedSlices)
	}
	if progress.Stories[1].Slices[0].RefactorHint != "extract helper" {
		t.Fatalf("RunProgress().Stories[1].Slices[0].RefactorHint = %q, want extract helper", progress.Stories[1].Slices[0].RefactorHint)
	}
}

func TestRunProgressNilPRD(t *testing.T) {
	var p *PRD
	if got := p.RunProgress(); got != nil {
		t.Fatalf("RunProgress() = %+v, want nil", got)
	}
}
