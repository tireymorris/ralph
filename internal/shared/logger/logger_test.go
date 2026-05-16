package logger

import (
	"testing"
)

func TestGetInitializesDefault(t *testing.T) {

	defaultLogger = nil

	Info("auto init test")

	if defaultLogger == nil {
		t.Error("defaultLogger should be auto-initialized")
	}
}
