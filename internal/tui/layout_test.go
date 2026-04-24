package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/config"
)

func TestApplyLayoutSetsPaneDimensions(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	m.applyLayout(120, 40)

	if m.mainPane.Width <= 0 {
		t.Errorf("main pane width should be positive, got %d", m.mainPane.Width)
	}
	if m.mainPane.Height <= 0 {
		t.Errorf("main pane height should be positive, got %d", m.mainPane.Height)
	}
	if m.logger.logView.Width <= 0 {
		t.Errorf("log view width should be positive, got %d", m.logger.logView.Width)
	}
	if m.logger.logView.Height <= 0 {
		t.Errorf("log view height should be positive, got %d", m.logger.logView.Height)
	}
}

func TestApplyLayoutCachesDimensions(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	m.applyLayout(120, 40)
	origMainH := m.mainPane.Height

	// Calling again with same dimensions should not change anything
	m.applyLayout(120, 40)
	if m.mainPane.Height != origMainH {
		t.Error("applyLayout should cache dimensions")
	}
}

func TestTabSwitchesPaneFocus(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhaseImplementation
	m.scrollPane = focusMain

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})

	if model, ok := newModel.(*Model); ok {
		if model.scrollPane != focusLogs {
			t.Errorf("scrollPane = %v, want focusLogs after Tab", model.scrollPane)
		}
	}
}

func TestTabSwitchesPaneFocusBack(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhaseImplementation
	m.scrollPane = focusLogs

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})

	if model, ok := newModel.(*Model); ok {
		if model.scrollPane != focusMain {
			t.Errorf("scrollPane = %v, want focusMain after Tab", model.scrollPane)
		}
	}
}
