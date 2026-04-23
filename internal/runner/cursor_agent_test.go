package runner

import (
	"testing"

	"ralph/internal/config"
)

func TestNewCursorAgent(t *testing.T) {
	cfg := &config.Config{Model: "cursor-agent/sonnet-4"}
	r := NewCursorAgent(cfg)

	if r == nil {
		t.Fatal("NewCursorAgent() returned nil")
	}
	if r.CmdFunc == nil {
		t.Error("CmdFunc should not be nil")
	}
}
