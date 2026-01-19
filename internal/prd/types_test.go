package prd

import "testing"

func TestNextPendingStory(t *testing.T) {
	tests := []struct {
		name       string
		prd        *PRD
		maxRetries int
		wantID     string
	}{
		{
			name:       "empty stories",
			prd:        &PRD{Stories: []*Story{}},
			maxRetries: 3,
			wantID:     "",
		},
		{
			name: "all completed",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: true, Priority: 1},
				{ID: "2", Passes: true, Priority: 2},
			}},
			maxRetries: 3,
			wantID:     "",
		},
		{
			name: "all exceeded retries",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: false, RetryCount: 3, Priority: 1},
				{ID: "2", Passes: false, RetryCount: 3, Priority: 2},
			}},
			maxRetries: 3,
			wantID:     "",
		},
		{
			name: "returns lowest priority pending",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: false, Priority: 3},
				{ID: "2", Passes: false, Priority: 1},
				{ID: "3", Passes: false, Priority: 2},
			}},
			maxRetries: 3,
			wantID:     "2",
		},
		{
			name: "skips completed stories",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: true, Priority: 1},
				{ID: "2", Passes: false, Priority: 2},
			}},
			maxRetries: 3,
			wantID:     "2",
		},
		{
			name: "skips exceeded retry stories",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: false, RetryCount: 5, Priority: 1},
				{ID: "2", Passes: false, RetryCount: 1, Priority: 2},
			}},
			maxRetries: 3,
			wantID:     "2",
		},
		{
			name: "respects maxRetries boundary",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: false, RetryCount: 2, Priority: 1},
			}},
			maxRetries: 3,
			wantID:     "1",
		},
		{
			name: "at maxRetries is excluded",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: false, RetryCount: 3, Priority: 1},
			}},
			maxRetries: 3,
			wantID:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.prd.NextPendingStory(tt.maxRetries)
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

func TestFailedStories(t *testing.T) {
	tests := []struct {
		name       string
		prd        *PRD
		maxRetries int
		wantIDs    []string
	}{
		{
			name:       "empty stories",
			prd:        &PRD{Stories: []*Story{}},
			maxRetries: 3,
			wantIDs:    nil,
		},
		{
			name: "no failed",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: true},
				{ID: "2", Passes: false, RetryCount: 1},
			}},
			maxRetries: 3,
			wantIDs:    nil,
		},
		{
			name: "some failed",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: false, RetryCount: 3},
				{ID: "2", Passes: true},
				{ID: "3", Passes: false, RetryCount: 5},
			}},
			maxRetries: 3,
			wantIDs:    []string{"1", "3"},
		},
		{
			name: "completed not counted as failed",
			prd: &PRD{Stories: []*Story{
				{ID: "1", Passes: true, RetryCount: 5},
			}},
			maxRetries: 3,
			wantIDs:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.prd.FailedStories(tt.maxRetries)
			if len(got) != len(tt.wantIDs) {
				t.Errorf("FailedStories() returned %d stories, want %d", len(got), len(tt.wantIDs))
				return
			}
			for i, story := range got {
				if story.ID != tt.wantIDs[i] {
					t.Errorf("FailedStories()[%d].ID = %q, want %q", i, story.ID, tt.wantIDs[i])
				}
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
