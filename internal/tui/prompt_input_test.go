package tui

import (
	"testing"

	"ralph/internal/shared/config"
)

func TestPromptInputPlaceholder(t *testing.T) {
	m := NewModel(config.DefaultConfig(), "", false, false, false)
	if m.promptInput.Placeholder == "" {
		t.Error("promptInput.Placeholder should be set for empty-prompt model")
	}
}

func TestPromptInputCharLimit(t *testing.T) {
	m := NewModel(config.DefaultConfig(), "", false, false, false)
	if m.promptInput.CharLimit < 2000 {
		t.Errorf("promptInput.CharLimit = %d, want at least 2000", m.promptInput.CharLimit)
	}
}
