package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runstate"
	"ralph/internal/shared/session"
	"ralph/internal/workflow"
)

func TestWaitingCleanupReviewFromCheckpoint(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	writeImplReviewCheckpoint(t, cfg.WorkDir)

	om := NewOperationManager(cfg)
	om.SetReviewLoop(runstate.LocalRunID, workflow.NewFileReviewLoop(cfg.WorkDir, runstate.LocalRunID))

	m := NewModel(cfg, "goal", false, false, false)
	m.phase = PhaseCleanup
	m.operationManager = om
	m.prd = completedPRDForCleanupReview()

	if !m.waitingCleanupReview() {
		t.Fatal("waitingCleanupReview() = false, want true from impl_review checkpoint")
	}
}

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

func TestUpdatePhaseCleanupEnterContinuesFromCheckpoint(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	writeImplReviewCheckpoint(t, cfg.WorkDir)

	om := NewOperationManager(cfg)
	om.SetReviewLoop(runstate.LocalRunID, workflow.NewFileReviewLoop(cfg.WorkDir, runstate.LocalRunID))

	m := NewModel(cfg, "goal", false, false, false)
	m.phase = PhaseCleanup
	m.operationManager = om
	m.prd = completedPRDForCleanupReview()

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(*Model)
	if model.phase != PhaseCleanup {
		t.Fatalf("phase = %v, want PhaseCleanup", model.phase)
	}
	if !assertContinueImplementationReviewCmd(t, cmd) {
		t.Fatal("expected ContinueImplementationReview delegation from checkpoint")
	}
}

func writeImplReviewCheckpoint(t *testing.T, workDir string) {
	t.Helper()
	metaDir := filepath.Join(workDir, ".ralph", "runs", runstate.LocalRunID)
	if err := os.MkdirAll(metaDir, 0o750); err != nil {
		t.Fatal(err)
	}
	meta := map[string]any{
		"checkpoint":       runstate.CheckpointImplReview,
		"review_iteration": 1,
	}
	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(metaDir, "meta.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func completedPRDForCleanupReview() *prd.PRD {
	return &prd.PRD{
		Stories: []*prd.Story{{
			ID: "s1", Title: "Story", Description: "d", Priority: 1, Passes: true,
			Slices: []*prd.Slice{{ID: "slice-1", Behavior: "AC", RedHint: "test", Passes: true}},
		}},
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
