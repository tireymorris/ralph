package workflow

import (
	"testing"

	"ralph/internal/shared/prd"
)

func TestSaveSingleStoryPRDSeedsSlices(t *testing.T) {
	cfg, saved := saveSingleStoryPRD(t, false)

	if len(saved.Stories) != 1 {
		t.Fatalf("Stories len = %d, want 1", len(saved.Stories))
	}
	slices := saved.Stories[0].Slices
	if len(slices) != 1 {
		t.Fatalf("Slices len = %d, want 1", len(slices))
	}
	if slices[0].ID != "slice-1" || slices[0].Behavior != "AC" || slices[0].RedHint == "" {
		t.Fatalf("slice = %+v, want id slice-1 behavior AC with non-empty red_hint", slices[0])
	}

	loaded, err := prd.Load(cfg)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Stories[0].Slices) != 1 {
		t.Fatalf("loaded Slices len = %d, want 1", len(loaded.Stories[0].Slices))
	}
}
