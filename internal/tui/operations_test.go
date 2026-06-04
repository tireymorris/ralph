package tui

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/clean"
	"ralph/internal/shared/config"
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
