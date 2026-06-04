package runs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"ralph/internal/shared/config"
)

func TestOngoingLocalPRD_incompletePRD(t *testing.T) {
	workDir := t.TempDir()
	prdJSON := `{
  "version": 1,
  "project_name": "My CLI project",
  "stories": [
    {"id": "s1", "title": "a", "description": "d", "acceptance_criteria": ["c"], "priority": 1, "passes": false}
  ]
}`
	if err := os.WriteFile(filepath.Join(workDir, "prd.json"), []byte(prdJSON), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	reg := NewRegistry()

	run, ok := OngoingLocalPRD(cfg, reg)
	if !ok {
		t.Fatal("OngoingLocalPRD() = false, want true")
	}
	if run.ID != LocalPRDRunID {
		t.Fatalf("ID = %q, want %q", run.ID, LocalPRDRunID)
	}
	if run.Prompt != "My CLI project" {
		t.Fatalf("Prompt = %q, want %q", run.Prompt, "My CLI project")
	}
	if run.Status != "implementing" {
		t.Fatalf("Status = %q, want implementing", run.Status)
	}
}

func TestOngoingLocalPRD_skipsWhenActiveWebRun(t *testing.T) {
	workDir := t.TempDir()
	prdJSON := `{
  "version": 1,
  "project_name": "x",
  "stories": [
    {"id": "s1", "title": "a", "description": "d", "acceptance_criteria": ["c"], "priority": 1, "passes": false}
  ]
}`
	if err := os.WriteFile(filepath.Join(workDir, "prd.json"), []byte(prdJSON), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	reg := NewRegistry()
	now := time.Now()
	if err := reg.Register(&Run{
		ID:        "web-run",
		WorkDir:   workDir,
		Prompt:    "web",
		Status:    "implementing",
		Phase:     "implement",
		CreatedAt: now,
		UpdatedAt: now,
		PRDPath:   "prd.json",
	}); err != nil {
		t.Fatal(err)
	}

	if _, ok := OngoingLocalPRD(cfg, reg); ok {
		t.Fatal("OngoingLocalPRD() = true, want false when web run is active")
	}
}

func TestOngoingLocalPRD_skipsCompleted(t *testing.T) {
	workDir := t.TempDir()
	prdJSON := `{
  "version": 1,
  "project_name": "done",
  "stories": [
    {"id": "s1", "title": "a", "description": "d", "acceptance_criteria": ["c"], "priority": 1, "passes": true}
  ]
}`
	if err := os.WriteFile(filepath.Join(workDir, "prd.json"), []byte(prdJSON), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir

	if _, ok := OngoingLocalPRD(cfg, NewRegistry()); ok {
		t.Fatal("OngoingLocalPRD() = true, want false for completed PRD")
	}
}

func TestOngoingLocalPRD_missingFile(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	if _, ok := OngoingLocalPRD(cfg, NewRegistry()); ok {
		t.Fatal("OngoingLocalPRD() = true, want false when prd.json missing")
	}
}
