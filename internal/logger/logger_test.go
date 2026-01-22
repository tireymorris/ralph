package logger

import (
	"testing"
)

func TestGetInitializesDefault(t *testing.T) {
	// Reset the defaultLogger to test auto-initialization
	defaultLogger = nil

	// This should auto-initialize
	Info("auto init test")

	// Should not panic and defaultLogger should be set
	if defaultLogger == nil {
		t.Error("defaultLogger should be auto-initialized")
	}
}
