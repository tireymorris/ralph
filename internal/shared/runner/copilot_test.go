package runner

import (
	"testing"

	"ralph/internal/shared/config"
)

func TestNewCopilot(t *testing.T) {
	cfg := &config.Config{Runner: "copilot"}
	r := NewCopilot(cfg)

	if r == nil {
		t.Fatal("NewCopilot() returned nil")
	}
	if r.CmdFunc == nil {
		t.Error("CmdFunc should not be nil")
	}
}
