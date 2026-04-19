package tui

import (
	"strings"
	"testing"
)

func TestGetStatusIcon(t *testing.T) {
	tests := []struct {
		name       string
		passes     bool
		inProgress bool
		wantIcon   string
	}{
		{
			name:       "completed",
			passes:     true,
			inProgress: false,
			wantIcon:   iconCompleted,
		},
		{
			name:       "in progress",
			passes:     false,
			inProgress: true,
			wantIcon:   iconInProgress,
		},
		{
			name:       "pending",
			passes:     false,
			inProgress: false,
			wantIcon:   iconPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStatusIcon(tt.passes, tt.inProgress)
			if !strings.Contains(got, tt.wantIcon) {
				t.Errorf("getStatusIcon() = %q, want to contain %q", got, tt.wantIcon)
			}
		})
	}
}

func TestGetStatusText(t *testing.T) {
	tests := []struct {
		name       string
		passes     bool
		inProgress bool
		wantText   string
	}{
		{
			name:       "completed",
			passes:     true,
			inProgress: false,
			wantText:   "completed",
		},
		{
			name:       "in progress",
			passes:     false,
			inProgress: true,
			wantText:   "in progress",
		},
		{
			name:       "pending",
			passes:     false,
			inProgress: false,
			wantText:   "pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStatusText(tt.passes, tt.inProgress)
			if !strings.Contains(got, tt.wantText) {
				t.Errorf("getStatusText() = %q, want to contain %q", got, tt.wantText)
			}
		})
	}
}
