package session

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
	"ralph/internal/shared/runstate"
	"ralph/internal/workflow"
	"ralph/internal/workflow/events"
)

type noopRunner struct{}

func (noopRunner) Run(context.Context, string, chan<- runner.OutputLine) error { return nil }
func (noopRunner) RunnerName() string                                          { return "noop" }
func (noopRunner) CommandName() string                                         { return "noop" }
func (noopRunner) IsInternalLog(string) bool                                   { return false }

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "t@example.com"},
		{"config", "user.name", "t"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func TestContinueImplementationReviewFromPRDDelegatesToDriver(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	cfg.PRDFile = "prd.json"
	cfg.SkipCleanup = true
	initGitRepo(t, cfg.WorkDir)

	p := &prd.PRD{
		ProjectName: "Done",
		Stories: []*prd.Story{
			{ID: "1", Title: "Only", Priority: 1, Passes: true},
		},
	}
	if err := prd.Save(cfg, p); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	s := NewWithRunner(cfg, noopRunner{})
	s.ContinueImplementationReviewFromPRD(context.Background(), p)

	deadline := time.After(2 * time.Second)
	for {
		select {
		case evt, ok := <-s.EventsCh():
			if !ok {
				return
			}
			if _, done := evt.(events.EventCompleted); done {
				return
			}
		case <-deadline:
			t.Fatal("timed out waiting for implementation review continue to finish")
		}
	}
}

func TestPRDForImplementationLoadsFromDiskWhenNotInMemory(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	onDisk := &prd.PRD{
		ProjectName: "On disk",
		Stories: []*prd.Story{
			{ID: "disk", Title: "Disk story", Priority: 1},
		},
	}
	if err := prd.Save(cfg, onDisk); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	s := New(cfg)

	loaded, err := s.PRDForImplementation(cfg)
	if err != nil {
		t.Fatalf("PRDForImplementation() error = %v", err)
	}
	if loaded.ProjectName != "On disk" {
		t.Fatalf("ProjectName = %q, want %q", loaded.ProjectName, "On disk")
	}
}

func TestPRDForImplementationLoadsMissingPRDWithImplementationError(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	s := New(cfg)

	_, err := s.PRDForImplementation(cfg)

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "load PRD for implementation") {
		t.Fatalf("expected load PRD for implementation error, got %q", err.Error())
	}
}

func TestApproveReviewLoadsMissingPRDWithImplementationError(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	s := New(cfg)

	err := s.ApproveReview(context.Background(), cfg)

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "load PRD for implementation") {
		t.Fatalf("expected load PRD for implementation error, got %q", err.Error())
	}
}

func TestResetPRDForImplementationUnmarksAndReloadsPRD(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	original := &prd.PRD{
		ProjectName: "Reset",
		Stories: []*prd.Story{
			{ID: "done", Title: "Done", Priority: 1, Passes: true},
		},
	}
	if err := prd.Save(cfg, original); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	s := New(cfg)
	reset, err := s.ResetPRDForImplementation(cfg)
	if err != nil {
		t.Fatalf("ResetPRDForImplementation() error = %v", err)
	}
	if reset.Stories[0].Passes {
		t.Fatal("ResetPRDForImplementation() returned story with Passes=true")
	}

	reloaded, err := prd.Load(cfg)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if reloaded.Stories[0].Passes {
		t.Fatal("ResetPRDForImplementation() saved story with Passes=true")
	}
}

func TestRunSnapshotBeforePRDLoadUsesPromptAndFallbackPhase(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	metaDir := filepath.Join(cfg.WorkDir, ".ralph", "runs", runstate.LocalRunID)
	if err := os.MkdirAll(metaDir, 0o750); err != nil {
		t.Fatalf("mkdir meta dir: %v", err)
	}
	meta, err := json.Marshal(map[string]string{"checkpoint": runstate.CheckpointFollowup})
	if err != nil {
		t.Fatalf("marshal meta: %v", err)
	}
	if err := os.WriteFile(filepath.Join(metaDir, "meta.json"), meta, 0o600); err != nil {
		t.Fatalf("write meta: %v", err)
	}

	s := NewWithRunner(cfg, noopRunner{})
	t.Cleanup(func() {
		s.Cancel()
		s.Wait()
	})
	s.SetReviewLoop(runstate.LocalRunID, workflow.NewFileReviewLoop(cfg.WorkDir, runstate.LocalRunID))
	s.StartNew(context.Background(), "build something")

	snap := s.RunSnapshot(runstate.PhaseImplement)

	if snap.Prompt != "build something" {
		t.Fatalf("Prompt = %q, want %q", snap.Prompt, "build something")
	}
	if snap.CurrentPRD != nil {
		t.Fatalf("CurrentPRD = %#v, want nil before PRD load", snap.CurrentPRD)
	}
	if snap.Phase != runstate.PhaseFollowup {
		t.Fatalf("Phase = %q, want %q", snap.Phase, runstate.PhaseFollowup)
	}
}
