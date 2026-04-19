package prd

import "testing"

func TestNextPendingStory(t *testing.T) {
	tests := []struct {
		name   string
		prd    *PRD
		wantID string
	}{
		{
			name:   "empty stories",
			prd:    &PRD{Stories: []*Story{}},
			wantID: "",
		},
		{
			name: "all completed",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: true, Priority: 1},
				{ID: "2", Passes: true, Priority: 2},
			}},
			wantID: "",
		},
		{
			name: "returns lowest priority pending",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: false, Priority: 3},
				{ID: "2", Passes: false, Priority: 1},
				{ID: "3", Passes: false, Priority: 2},
			}},
			wantID: "2",
		},
		{
			name: "skips completed stories",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: true, Priority: 1},
				{ID: "2", Passes: false, Priority: 2},
			}},
			wantID: "2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.prd.NextPendingStory()
			if tt.wantID == "" {
				if got != nil {
					t.Errorf("NextPendingStory() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("NextPendingStory() = nil, want ID %q", tt.wantID)
				} else if got.ID != tt.wantID {
					t.Errorf("NextPendingStory().ID = %q, want %q", got.ID, tt.wantID)
				}
			}
		})
	}
}

func TestCompletedCount(t *testing.T) {
	tests := []struct {
		name string
		prd  *PRD
		want int
	}{
		{
			name: "empty stories",
			prd:  &PRD{Stories: []*Story{}},
			want: 0,
		},
		{
			name: "no completed",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: false},
				{ID: "2", Passes: false},
			}},
			want: 0,
		},
		{
			name: "all completed",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: true},
				{ID: "2", Passes: true},
			}},
			want: 2,
		},
		{
			name: "mixed",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: true},
				{ID: "2", Passes: false},
				{ID: "3", Passes: true},
			}},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.prd.CompletedCount()
			if got != tt.want {
				t.Errorf("CompletedCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestAllCompleted(t *testing.T) {
	tests := []struct {
		name string
		prd  *PRD
		want bool
	}{
		{
			name: "empty stories",
			prd:  &PRD{Stories: []*Story{}},
			want: true,
		},
		{
			name: "all completed",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: true},
				{ID: "2", Passes: true},
			}},
			want: true,
		},
		{
			name: "none completed",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: false},
				{ID: "2", Passes: false},
			}},
			want: false,
		},
		{
			name: "partial completed",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: true},
				{ID: "2", Passes: false},
			}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.prd.AllCompleted()
			if got != tt.want {
				t.Errorf("AllCompleted() = %v, want %v", got, tt.want)
			}
		})
	}
}
