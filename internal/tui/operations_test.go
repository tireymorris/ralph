package tui

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/clean"
	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runstate"
)

func TestStartFullOperationNonResumeArchivesPriorPRD(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir

	if _, err := clean.SeedStateArtifacts(cfg); err != nil {
		t.Fatalf("SeedStateArtifacts: %v", err)
	}
	if _, err := os.Stat(cfg.PRDPath()); err != nil {
		t.Fatalf("seeded prd.json: %v", err)
	}

	om := NewOperationManager(cfg)
	defer om.Cancel()

	_ = om.StartFullOperation(false, "new goal")()

	if _, err := os.Stat(cfg.PRDPath()); !os.IsNotExist(err) {
		t.Fatalf("prd.json should be absent after non-resume start, stat err=%v", err)
	}

	backups, err := filepath.Glob(filepath.Join(workDir, ".ralph", "backups", "*"))
	if err != nil {
		t.Fatalf("glob backups: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("expected one backup dir, got %d: %v", len(backups), backups)
	}
	if _, err := os.Stat(filepath.Join(backups[0], "prd.json")); err != nil {
		t.Fatalf("archived prd.json: %v", err)
	}
}

func TestStartFullOperationResumeSkipsArchive(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir

	if _, err := clean.SeedStateArtifacts(cfg); err != nil {
		t.Fatalf("SeedStateArtifacts: %v", err)
	}

	om := NewOperationManager(cfg)
	defer om.Cancel()

	_ = om.StartFullOperation(true, "")()

	if _, err := os.Stat(cfg.PRDPath()); err != nil {
		t.Fatalf("prd.json should remain on resume: %v", err)
	}

	backups, err := filepath.Glob(filepath.Join(workDir, ".ralph", "backups", "*"))
	if err != nil {
		t.Fatalf("glob backups: %v", err)
	}
	if len(backups) != 0 {
		t.Fatalf("resume should not create backups, got %v", backups)
	}
}

func TestResumeStartMsgImplementationPhase(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir

	p := &prd.PRD{
		ProjectName: "Resume Test",
		Stories: []*prd.Story{
			{ID: "1", Title: "Done", Passes: true, Priority: 1},
			{ID: "2", Title: "Next", Passes: false, Priority: 2},
		},
	}
	if err := prd.Save(cfg, p); err != nil {
		t.Fatalf("Save PRD: %v", err)
	}

	metaDir := filepath.Join(workDir, ".ralph", "runs", runstate.LocalRunID)
	if err := os.MkdirAll(metaDir, 0755); err != nil {
		t.Fatalf("mkdir meta: %v", err)
	}
	meta, _ := json.Marshal(map[string]string{"checkpoint": runstate.CheckpointFollowup})
	if err := os.WriteFile(filepath.Join(metaDir, "meta.json"), meta, 0644); err != nil {
		t.Fatalf("write meta: %v", err)
	}

	om := NewOperationManager(cfg)
	defer om.Cancel()

	msg := om.resumeStartMsg()
	rsm, ok := msg.(resumeStartMsg)
	if !ok {
		t.Fatalf("resumeStartMsg() = %T, want resumeStartMsg", msg)
	}
	if rsm.phase != PhaseImplementation {
		t.Errorf("phase = %v, want PhaseImplementation", rsm.phase)
	}
	if rsm.prd == nil || rsm.prd.ProjectName != "Resume Test" {
		t.Errorf("prd = %v, want Resume Test project", rsm.prd)
	}
}

func TestImplementationReviewEnterAttemptsMissingPRD(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	m := NewModel(cfg, "goal", false, false, false)
	m.phase = PhaseImplementationReview
	m.prd = nil

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command to report missing PRD")
	}
}

func TestContinueImplementationReviewReportsMissingPRD(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	om := NewOperationManager(cfg)
	defer om.Cancel()

	msg := om.ContinueImplementationReview()()
	errMsg, ok := msg.(operationErrorMsg)
	if !ok {
		t.Fatalf("ContinueImplementationReview() msg = %T, want operationErrorMsg", msg)
	}
	if errMsg.err == nil || !strings.Contains(errMsg.err.Error(), "load PRD for implementation") {
		t.Fatalf("error = %v, want load PRD for implementation", errMsg.err)
	}
}

func TestStartImplementationReportsMissingPRD(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	om := NewOperationManager(cfg)
	defer om.Cancel()

	msg := om.StartImplementation(nil)()
	errMsg, ok := msg.(operationErrorMsg)
	if !ok {
		t.Fatalf("StartImplementation(nil) msg = %T, want operationErrorMsg", msg)
	}
	if errMsg.err == nil || !strings.Contains(errMsg.err.Error(), "load PRD for implementation") {
		t.Fatalf("error = %v, want load PRD for implementation", errMsg.err)
	}
}

func TestUpdateOperationErrorMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "goal", false, false, false)

	newModel, _ := m.Update(operationErrorMsg{err: errors.New("archive prior state: denied")})
	model := newModel.(*Model)

	if model.phase != PhaseFailed {
		t.Errorf("phase = %v, want PhaseFailed", model.phase)
	}
	if model.err == nil || model.err.Error() != "archive prior state: denied" {
		t.Errorf("err = %v, want archive error", model.err)
	}
}
