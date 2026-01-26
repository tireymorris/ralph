package prd

import (
	"strings"
	"testing"
)

func TestPRD_Validate(t *testing.T) {
	tests := []struct {
		name    string
		prd     *PRD
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid PRD",
			prd: &PRD{
				ProjectName: "Test Project",
				Context:     "Valid context",
				Stories: []*Story{
					{ID: "story-1", Title: "Story 1", Description: "Description", Priority: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "context too large",
			prd: &PRD{
				ProjectName: "Test Project",
				Context:     string(make([]byte, MaxContextSize+1)),
				Stories:     []*Story{},
			},
			wantErr: true,
			errMsg:  "context size",
		},
		{
			name: "too many stories",
			prd: &PRD{
				ProjectName: "Test Project",
				Stories:     make([]*Story, MaxStories+1),
			},
			wantErr: true,
			errMsg:  "story count",
		},
		{
			name: "duplicate story IDs",
			prd: &PRD{
				ProjectName: "Test Project",
				Stories: []*Story{
					{ID: "story-1", Title: "Story 1", Description: "Description", Priority: 1},
					{ID: "story-1", Title: "Story 2", Description: "Description", Priority: 2},
				},
			},
			wantErr: true,
			errMsg:  "duplicate story ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.prd.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PRD.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("PRD.Validate() error = %v, expected to contain %q", err, tt.errMsg)
			}
		})
	}
}

func TestStory_Validate(t *testing.T) {
	tests := []struct {
		name    string
		story   *Story
		seenIDs map[string]bool
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid story",
			story: &Story{
				ID:          "story-1",
				Title:       "Story 1",
				Description: "Valid description",
				Priority:    1,
			},
			seenIDs: make(map[string]bool),
			wantErr: false,
		},
		{
			name:    "empty ID",
			story:   &Story{Title: "Story 1", Description: "Description", Priority: 1},
			seenIDs: make(map[string]bool),
			wantErr: true,
			errMsg:  "story ID cannot be empty",
		},
		{
			name:    "duplicate ID",
			story:   &Story{ID: "story-1", Title: "Story 1", Description: "Description", Priority: 1},
			seenIDs: map[string]bool{"story-1": true},
			wantErr: true,
			errMsg:  "duplicate story ID",
		},
		{
			name:    "empty title",
			story:   &Story{ID: "story-1", Description: "Description", Priority: 1},
			seenIDs: make(map[string]bool),
			wantErr: true,
			errMsg:  "story title cannot be empty",
		},
		{
			name:    "description too large",
			story:   &Story{ID: "story-1", Title: "Story 1", Description: string(make([]byte, MaxStoryDescSize+1)), Priority: 1},
			seenIDs: make(map[string]bool),
			wantErr: true,
			errMsg:  "story description size",
		},
		{
			name:    "negative priority",
			story:   &Story{ID: "story-1", Title: "Story 1", Description: "Description", Priority: -1},
			seenIDs: make(map[string]bool),
			wantErr: true,
			errMsg:  "story priority",
		},
		{
			name:    "too many acceptance criteria",
			story:   &Story{ID: "story-1", Title: "Story 1", Description: "Description", AcceptanceCriteria: make([]string, MaxAcceptanceCriteria+1)},
			seenIDs: make(map[string]bool),
			wantErr: true,
			errMsg:  "acceptance criteria",
		},
		{
			name:    "negative retry count",
			story:   &Story{ID: "story-1", Title: "Story 1", Description: "Description", RetryCount: -1},
			seenIDs: make(map[string]bool),
			wantErr: true,
			errMsg:  "story retry count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.story.Validate(tt.seenIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Story.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Story.Validate() error = %v, expected to contain %q", err, tt.errMsg)
			}
		})
	}
}
