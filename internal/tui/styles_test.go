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
		retryCount int
		maxRetries int
		wantIcon   string
	}{
		{
			name:       "completed",
			passes:     true,
			inProgress: false,
			retryCount: 0,
			maxRetries: 3,
			wantIcon:   iconCompleted,
		},
		{
			name:       "in progress",
			passes:     false,
			inProgress: true,
			retryCount: 0,
			maxRetries: 3,
			wantIcon:   iconInProgress,
		},
		{
			name:       "failed max retries",
			passes:     false,
			inProgress: false,
			retryCount: 3,
			maxRetries: 3,
			wantIcon:   iconFailed,
		},
		{
			name:       "failed exceeded retries",
			passes:     false,
			inProgress: false,
			retryCount: 5,
			maxRetries: 3,
			wantIcon:   iconFailed,
		},
		{
			name:       "pending",
			passes:     false,
			inProgress: false,
			retryCount: 0,
			maxRetries: 3,
			wantIcon:   iconPending,
		},
		{
			name:       "pending with some retries",
			passes:     false,
			inProgress: false,
			retryCount: 2,
			maxRetries: 3,
			wantIcon:   iconPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStatusIcon(tt.passes, tt.inProgress, tt.retryCount, tt.maxRetries)
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
		retryCount int
		maxRetries int
		wantText   string
	}{
		{
			name:       "completed",
			passes:     true,
			inProgress: false,
			retryCount: 0,
			maxRetries: 3,
			wantText:   "completed",
		},
		{
			name:       "in progress",
			passes:     false,
			inProgress: true,
			retryCount: 0,
			maxRetries: 3,
			wantText:   "in progress",
		},
		{
			name:       "failed",
			passes:     false,
			inProgress: false,
			retryCount: 3,
			maxRetries: 3,
			wantText:   "failed",
		},
		{
			name:       "pending",
			passes:     false,
			inProgress: false,
			retryCount: 0,
			maxRetries: 3,
			wantText:   "pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStatusText(tt.passes, tt.inProgress, tt.retryCount, tt.maxRetries)
			if !strings.Contains(got, tt.wantText) {
				t.Errorf("getStatusText() = %q, want to contain %q", got, tt.wantText)
			}
		})
	}
}
