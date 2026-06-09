package tui

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/clean"
	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runstate"
	"ralph/internal/shared/runner"
	"ralph/internal/shared/session"
	"ralph/internal/workflow/events"
)

type noopRunner struct{}

func (noopRunner) Run(context.Context, string, chan<- runner.OutputLine) error { return nil }
func (noopRunner) RunnerName() string                                          { return "noop" }
func (noopRunner) CommandName() string                                         { return "noop" }
func (noopRunner) IsInternalLog(string) bool                                   { return false }

func waitSessionDone(t *testing.T, om *OperationManager) {
	t.Helper()
	deadline := time.After(3 * time.Second)
	for {
		select {
		case ev, ok := <-om.EventsCh():
			if !ok {
				om.Cancel()
				break
			}
			switch ev.(type) {
			case events.EventCompleted, events.EventError:
				om.Cancel()
				goto waitCtxDone
			}
		case <-deadline:
			om.Cancel()
			goto waitCtxDone
		}
	}
waitCtxDone:
	select {
	case <-om.Ctx().Done():
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for session cancellation")
	}
}

func TestStartFullOperationNonResumeArchivesPriorPRD(t *testing.T) {
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
	t.Cleanup(func() { waitSessionDone(t, om) })

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
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir

	if _, err := clean.SeedStateArtifacts(cfg); err != nil {
		t.Fatalf("SeedStateArtifacts: %v", err)
	}

	om := NewOperationManager(cfg)
	t.Cleanup(func() { waitSessionDone(t, om) })

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
	t.Cleanup(func() { waitSessionDone(t, om) })

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

func TestImplementationReviewEnterReportsMissingPRD(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	m := NewModel(cfg, "goal", false, false, false)
	m.phase = PhaseImplementationReview
	m.prd = nil

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command to report missing PRD")
	}
	model := updated.(*Model)
	if model.phase != PhaseImplementation {
		t.Fatalf("phase = %v, want PhaseImplementation", model.phase)
	}

	msg := m.operationManager.ContinueImplementationReview()()
	errMsg, ok := msg.(operationErrorMsg)
	if !ok {
		t.Fatalf("ContinueImplementationReview() msg = %T, want operationErrorMsg", msg)
	}
	if errMsg.err == nil || !strings.Contains(errMsg.err.Error(), "load PRD for implementation") {
		t.Fatalf("error = %v, want load PRD for implementation", errMsg.err)
	}
}

func TestPRDReviewEnterApprovesWithoutInMemoryPRD(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	cfg.SkipCleanup = true

	onDisk := &prd.PRD{
		ProjectName: "On disk",
		Stories: []*prd.Story{
			{ID: "disk", Title: "Disk story", Description: "from disk", Priority: 1},
		},
	}
	if err := prd.Save(cfg, onDisk); err != nil {
		t.Fatalf("Save PRD: %v", err)
	}

	m := NewModel(cfg, "goal", false, false, false)
	m.operationManager = &OperationManager{Session: session.NewWithRunner(cfg, noopRunner{}), cfg: cfg}
	m.phase = PhasePRDReview
	m.prd = nil

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected approve command")
	}
	model := updated.(*Model)
	if model.phase != PhaseImplementation {
		t.Fatalf("phase = %v, want PhaseImplementation", model.phase)
	}

	om := m.operationManager
	t.Cleanup(func() { waitSessionDone(t, om) })

	msg := om.ApproveReview()()
	if msg != nil {
		if errMsg, ok := msg.(operationErrorMsg); ok {
			t.Fatalf("ApproveReview() error = %v", errMsg.err)
		}
		t.Fatalf("ApproveReview() msg = %T, want nil", msg)
	}
}

func TestApproveReviewReportsMissingPRD(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	om := NewOperationManager(cfg)
	t.Cleanup(func() { waitSessionDone(t, om) })

	msg := om.ApproveReview()()
	errMsg, ok := msg.(operationErrorMsg)
	if !ok {
		t.Fatalf("ApproveReview() msg = %T, want operationErrorMsg", msg)
	}
	if errMsg.err == nil || !strings.Contains(errMsg.err.Error(), "load PRD for implementation") {
		t.Fatalf("error = %v, want load PRD for implementation", errMsg.err)
	}
}

func TestContinueImplementationReviewReportsMissingPRD(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	om := NewOperationManager(cfg)
	t.Cleanup(func() { waitSessionDone(t, om) })

	msg := om.ContinueImplementationReview()()
	errMsg, ok := msg.(operationErrorMsg)
	if !ok {
		t.Fatalf("ContinueImplementationReview() msg = %T, want operationErrorMsg", msg)
	}
	if errMsg.err == nil || !strings.Contains(errMsg.err.Error(), "load PRD for implementation") {
		t.Fatalf("error = %v, want load PRD for implementation", errMsg.err)
	}
}

func TestStartImplementationFromPRDUsesSuppliedPRD(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	cfg.SkipCleanup = true

	supplied := &prd.PRD{
		ProjectName: "Supplied",
		Stories: []*prd.Story{
			{ID: "supplied", Title: "Use supplied story", Description: "implement supplied story", Priority: 1},
		},
	}
	if err := prd.Save(cfg, supplied); err != nil {
		t.Fatalf("Save supplied PRD: %v", err)
	}

	current := &prd.PRD{
		ProjectName: "Current",
		Stories: []*prd.Story{
			{ID: "current", Title: "Current story", Description: "do not use", Priority: 1},
		},
	}

	om := &OperationManager{Session: session.NewWithRunner(cfg, noopRunner{}), cfg: cfg}
	om.TrackEventState(events.EventPRDGenerated{PRD: current})

	msg := om.StartImplementation(supplied)()
	if msg != nil {
		t.Fatalf("StartImplementation(supplied) msg = %T, want nil", msg)
	}

	waitForStoryStarted(t, om, "supplied")
	waitSessionDone(t, om)
}

func waitForStoryStarted(t *testing.T, om *OperationManager, wantID string) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case ev := <-om.EventsCh():
			switch e := ev.(type) {
			case events.EventStoryStarted:
				if e.Story.ID != wantID {
					t.Fatalf("started story ID = %q, want %q", e.Story.ID, wantID)
				}
				return
			case events.EventError:
				t.Fatalf("unexpected error event: %v", e.Err)
			}
		case <-deadline:
			t.Fatal("timed out waiting for implementation to start")
		}
	}
}

func TestStartImplementationReportsMissingPRD(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	om := NewOperationManager(cfg)
	t.Cleanup(func() { waitSessionDone(t, om) })

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
