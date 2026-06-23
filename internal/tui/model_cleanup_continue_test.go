package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/shared/config"
	"ralph/internal/shared/session"
)

func TestUpdatePhaseCleanupEnterContinuesImplementationReview(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	m := NewModel(cfg, "goal", false, false, false)
	m.phase = PhaseCleanup
	m.activity = session.RunActivity{
		Kind:         session.ActivityReview,
		FindingCount: 2,
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(*Model)
	if model.phase != PhaseCleanup {
		t.Fatalf("phase = %v, want PhaseCleanup", model.phase)
	}
	if !assertContinueImplementationReviewCmd(t, cmd) {
		t.Fatal("expected ContinueImplementationReview delegation")
	}
}

func assertContinueImplementationReviewCmd(t *testing.T, cmd tea.Cmd) bool {
	t.Helper()
	if cmd == nil {
		return false
	}
	msg := cmd()
	if errMsg, ok := msg.(operationErrorMsg); ok {
		return errMsg.err != nil && strings.Contains(errMsg.err.Error(), "load PRD for implementation")
	}
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		return false
	}
	for _, subCmd := range batch {
		if subCmd == nil {
			continue
		}
		subMsg := subCmd()
		if subMsg == nil {
			continue
		}
		if errMsg, ok := subMsg.(operationErrorMsg); ok {
			if errMsg.err != nil && strings.Contains(errMsg.err.Error(), "load PRD for implementation") {
				return true
			}
		}
	}
	return false
}
