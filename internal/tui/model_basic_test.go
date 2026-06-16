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
	"ralph/internal/workflow/events"
)

func TestPhaseString(t *testing.T) {
	tests := []struct {
		phase Phase
		want  string
	}{
		{PhaseInit, "Initializing"},
		{PhaseAwaitingPrompt, "Awaiting Prompt"},
		{PhaseClarifying, "Clarifying Questions"},
		{PhasePRDGeneration, "Phase 1: PRD Generation"},
		{PhasePRDReview, "PRD Review"},
		{PhaseImplementation, "Phase 2: Implementation"},
		{PhaseCleanup, "Phase 3: Cleanup"},
		{PhaseCompleted, "Completed"},
		{PhaseFailed, "Failed"},
		{Phase(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.phase.String()
			if got != tt.want {
				t.Errorf("Phase.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPhaseCleanupOrdering(t *testing.T) {
	if PhaseCleanup <= PhaseImplementation {
		t.Errorf("PhaseCleanup (%d) should be > PhaseImplementation (%d)", PhaseCleanup, PhaseImplementation)
	}
	if PhaseCleanup >= PhaseCompleted {
		t.Errorf("PhaseCleanup (%d) should be < PhaseCompleted (%d)", PhaseCleanup, PhaseCompleted)
	}
}

func TestNewModel(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test prompt", true, false, false)

	if m.cfg != cfg {
		t.Error("cfg not set correctly")
	}
	if m.prompt != "test prompt" {
		t.Errorf("prompt = %q, want %q", m.prompt, "test prompt")
	}
	if !m.dryRun {
		t.Error("dryRun should be true")
	}
	if m.resume {
		t.Error("resume should be false")
	}
	if m.phase != PhaseInit {
		t.Errorf("phase = %v, want PhaseInit", m.phase)
	}
	if m.operationManager == nil {
		t.Error("operationManager should not be nil")
	}
	if m.logger == nil {
		t.Error("logger should not be nil")
	}
}

func TestExitCode(t *testing.T) {
	tests := []struct {
		name     string
		phase    Phase
		prd      *prd.PRD
		wantCode int
	}{
		{
			name:     "completed",
			phase:    PhaseCompleted,
			prd:      nil,
			wantCode: 0,
		},
		{
			name:     "other phase",
			phase:    PhaseInit,
			prd:      nil,
			wantCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Model{phase: tt.phase, prd: tt.prd}
			got := m.ExitCode()
			if got != tt.wantCode {
				t.Errorf("ExitCode() = %d, want %d", got, tt.wantCode)
			}
		})
	}
}

func TestInit(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil")
	}
}

func TestInitEmptyPromptAwaitingPhase(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "", false, false, false)

	_ = m.Init()

	if m.phase != PhaseAwaitingPrompt {
		t.Errorf("phase = %v, want PhaseAwaitingPrompt", m.phase)
	}
}

func TestInitWithPromptStartsWorkflow(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "build api", false, false, false)

	_ = m.Init()

	if m.phase == PhaseAwaitingPrompt {
		t.Error("Init with prompt should not enter PhaseAwaitingPrompt")
	}
	if m.phase != PhaseInit {
		t.Errorf("phase = %v, want PhaseInit before workflow starts", m.phase)
	}

	msg := m.operationManager.StartFullOperation(m.resume, m.prompt)()
	newModel, _ := m.Update(msg)
	m = newModel.(*Model)

	if m.phase != PhasePRDGeneration {
		t.Errorf("phase = %v, want PhasePRDGeneration", m.phase)
	}
}

func TestInitResumeNeverPhaseAwaitingPrompt(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "", false, true, false)

	_ = m.Init()

	if m.phase == PhaseAwaitingPrompt {
		t.Errorf("phase = %v, resume with empty prompt must not enter PhaseAwaitingPrompt", m.phase)
	}
}

func TestInitResumeEmptyPromptStartsWorkflow(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "", false, true, false)

	_ = m.Init()

	if m.phase == PhaseAwaitingPrompt {
		t.Errorf("phase = %v, resume with empty prompt must not enter PhaseAwaitingPrompt", m.phase)
	}

	msg := m.operationManager.StartFullOperation(true, "")()
	newModel, _ := m.Update(msg)
	m = newModel.(*Model)

	if m.phase != PhasePRDGeneration {
		t.Errorf("phase = %v, want PhasePRDGeneration", m.phase)
	}
}

func TestResumeMainPaneShowsImplementationProgress(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.Runner = "mock"

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

	m := NewModel(cfg, "", false, true, false)
	t.Cleanup(func() { waitSessionDone(t, m.operationManager) })

	msg := m.operationManager.StartFullOperation(true, "")()
	newModel, _ := m.Update(msg)
	m = newModel.(*Model)

	if m.phase != PhaseImplementation {
		t.Fatalf("phase = %v, want PhaseImplementation", m.phase)
	}
	if m.prd == nil || m.prd.ProjectName != "Resume Test" {
		t.Fatalf("prd = %v, want Resume Test", m.prd)
	}

	m.width, m.height = 80, 45
	prepMainView(m)
	view := m.View()

	if strings.Contains(view, "Generating PRD") {
		t.Errorf("resume main pane should not show PRD generation, got %q", view)
	}
	if !strings.Contains(view, "1/2 stories") {
		t.Errorf("resume main pane should show story progress, got %q", view)
	}
	if !strings.Contains(view, "Phase 2: Implementation") {
		t.Errorf("resume main pane should show implementation phase, got %q", view)
	}
}

func TestResumeMainPaneShowsActiveStorySliceProgressFromSnapshot(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.Runner = "mock"

	p := &prd.PRD{
		ProjectName: "Resume Snapshot Test",
		Stories: []*prd.Story{
			{ID: "1", Title: "Done", Passes: true, Priority: 1},
			{
				ID:       "2",
				Title:    "Active",
				Passes:   false,
				Priority: 2,
				Slices: []*prd.Slice{
					{ID: "slice-1", Behavior: "first slice", RedHint: "write failing test", Passes: true},
					{ID: "slice-2", Behavior: "second slice", RedHint: "write failing test", Passes: false},
				},
			},
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

	m := NewModel(cfg, "", false, true, false)
	t.Cleanup(func() { waitSessionDone(t, m.operationManager) })

	msg := m.operationManager.StartFullOperation(true, "")()
	newModel, _ := m.Update(msg)
	m = newModel.(*Model)

	m.width, m.height = 80, 45
	prepMainView(m)
	view := m.View()

	for _, want := range []string{"first slice", "second slice", "completed", "in progress"} {
		if !strings.Contains(view, want) {
			t.Fatalf("resume main pane should show %q from the shared snapshot, got %q", want, view)
		}
	}
}

func awaitingPromptModel(t *testing.T) *Model {
	t.Helper()
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "", false, false, false)
	_ = m.Init()
	return m
}

func TestUpdateAwaitingPromptEnterSubmitPrompt(t *testing.T) {
	m := awaitingPromptModel(t)
	m.promptInput.SetValue("  build api  ")

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := newModel.(*Model)

	if model.prompt != "build api" {
		t.Errorf("prompt = %q, want %q", model.prompt, "build api")
	}
	if model.phase != PhasePRDGeneration {
		t.Errorf("phase = %v, want PhasePRDGeneration", model.phase)
	}
	if cmd == nil {
		t.Fatal("expected StartFullOperation cmd")
	}
}

func TestUpdateAwaitingPromptEnterWhitespaceOnly(t *testing.T) {
	m := awaitingPromptModel(t)
	m.promptInput.SetValue("   ")

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := newModel.(*Model)

	if model.prompt != "" {
		t.Errorf("prompt = %q, want empty", model.prompt)
	}
	if model.phase != PhaseAwaitingPrompt {
		t.Errorf("phase = %v, want PhaseAwaitingPrompt", model.phase)
	}
	if cmd != nil {
		t.Error("whitespace-only enter should not start workflow")
	}
}

func TestUpdateAwaitingPromptEnterStartsWorkflowOnce(t *testing.T) {
	m := awaitingPromptModel(t)
	m.promptInput.SetValue("build api")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected StartFullOperation cmd")
	}

	msg := cmd()
	pcm, ok := msg.(phaseChangeMsg)
	if !ok {
		t.Fatalf("cmd() = %T, want phaseChangeMsg", msg)
	}
	if Phase(pcm) != PhasePRDGeneration {
		t.Errorf("phaseChangeMsg = %v, want PhasePRDGeneration", pcm)
	}
}

func TestUpdateAwaitingPromptQuit(t *testing.T) {
	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyCtrlC},
		{Type: tea.KeyRunes, Runes: []rune{'q'}},
	} {
		m := awaitingPromptModel(t)
		newModel, cmd := m.Update(key)
		model := newModel.(*Model)

		if !model.quitting {
			t.Errorf("quitting should be true after %v from awaiting prompt", key)
		}
		if cmd == nil {
			t.Fatalf("expected tea.Quit after %v", key)
		}
	}
}

func TestUpdateKeyMsgQuit(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if model, ok := newModel.(*Model); ok {
		if !model.quitting {
			t.Error("quitting should be true after 'q' key")
		}
	}
	if cmd == nil {
		t.Error("should return quit command")
	}
}

func TestUpdateKeyMsgCtrlC(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if model, ok := newModel.(*Model); ok {
		if !model.quitting {
			t.Error("quitting should be true after Ctrl+C")
		}
	}
}

func TestUpdateWindowSizeMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if model, ok := newModel.(*Model); ok {
		if model.width != 120 {
			t.Errorf("width = %d, want 120", model.width)
		}
		if model.height != 40 {
			t.Errorf("height = %d, want 40", model.height)
		}
	}
}

func TestUpdatePRDGeneratedMsgDryRun(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", true, false, false)

	testPRD := &prd.PRD{ProjectName: "Test", Stories: []*prd.Story{{ID: "1"}}}
	newModel, _ := m.Update(workflowEventMsg{event: events.EventPRDGenerated{PRD: testPRD}})

	if model, ok := newModel.(*Model); ok {
		if model.prd != testPRD {
			t.Error("prd should be set")
		}
		if model.phase != PhaseCompleted {
			t.Errorf("phase = %v, want PhaseCompleted for dry run", model.phase)
		}
	}
}

func TestUpdatePRDGeneratedMsgDryRunTUIPromptSubmit(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "", true, false, false)
	_ = m.Init()
	m.promptInput.SetValue("build api")

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := newModel.(*Model)

	if model.prompt != "build api" {
		t.Fatalf("prompt = %q, want %q", model.prompt, "build api")
	}

	testPRD := &prd.PRD{ProjectName: "Test", Stories: []*prd.Story{{ID: "1"}}}
	newModel, _ = model.Update(workflowEventMsg{event: events.EventPRDGenerated{PRD: testPRD}})
	model = newModel.(*Model)

	if model.phase != PhaseCompleted {
		t.Errorf("phase = %v, want PhaseCompleted for dry run after TUI prompt", model.phase)
	}
}

func TestUpdatePRDGeneratedMsgImplement(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	testPRD := &prd.PRD{ProjectName: "Test", Stories: []*prd.Story{{ID: "1"}}}
	newModel, _ := m.Update(workflowEventMsg{event: events.EventPRDGenerated{PRD: testPRD}})

	if model, ok := newModel.(*Model); ok {
		if model.phase != PhasePRDReview {
			t.Errorf("phase = %v, want PhasePRDReview", model.phase)
		}
	}
}

func TestUpdatePRDErrorMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhasePRDGeneration

	testErr := &testErrorType{msg: "test error"}
	newModel, _ := m.Update(workflowEventMsg{event: events.EventError{Err: testErr}})

	if model, ok := newModel.(*Model); ok {
		if model.err != testErr {
			t.Error("err should be set")
		}
		if model.phase != PhaseFailed {
			t.Errorf("phase = %v, want PhaseFailed", model.phase)
		}
	}
}

func TestUpdateRetryAfterFailureRestartsPRDFlow(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "prompt", false, false, false)
	m.width, m.height = 120, 40
	m.phase = PhaseFailed
	m.err = &testErrorType{msg: "bad path"}
	m.retryImplementation = false

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model := newModel.(*Model)
	if model.err != nil {
		t.Error("err should clear on retry")
	}
	if model.phase != PhasePRDGeneration {
		t.Errorf("phase = %v, want PhasePRDGeneration", model.phase)
	}
	if cmd == nil {
		t.Error("expected retry command batch")
	}
}

func TestUpdateRetryAfterFailureResumesImplementation(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "prompt", false, false, false)
	m.width, m.height = 120, 40
	m.phase = PhaseFailed
	m.err = &testErrorType{msg: "story failed"}
	m.retryImplementation = true
	m.prd = &prd.PRD{ProjectName: "P"}

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model := newModel.(*Model)
	if model.err != nil {
		t.Error("err should clear on retry")
	}
	if model.phase != PhaseImplementation {
		t.Errorf("phase = %v, want PhaseImplementation", model.phase)
	}
	if cmd == nil {
		t.Error("expected retry command batch")
	}
}

func TestUpdatePhaseChangeMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	newModel, _ := m.Update(phaseChangeMsg(PhaseCompleted))

	if model, ok := newModel.(*Model); ok {
		if model.phase != PhaseCompleted {
			t.Errorf("phase = %v, want PhaseCompleted", model.phase)
		}
	}
}

func TestUpdateSpinnerTickMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	_, cmd := m.Update(m.spinner.Tick())
	if cmd == nil {
		t.Error("spinner tick should return a command")
	}
}

func TestUpdatePRDReviewCritiqueKeyOpensInputMode(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhasePRDReview
	m.prd = &prd.PRD{ProjectName: "P"}

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	model := newModel.(*Model)

	if !model.critiqueActive {
		t.Fatal("critiqueActive should be true after pressing critique key")
	}
	if model.phase != PhasePRDReview {
		t.Fatalf("phase = %v, want PhasePRDReview", model.phase)
	}
	if cmd == nil {
		t.Fatal("opening critique mode should return a command")
	}
}

func TestUpdatePRDReviewCritiqueEnterStartsRevision(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhasePRDReview
	m.prd = &prd.PRD{ProjectName: "P"}
	m.critiqueActive = true
	m.critiqueInput.SetValue("Needs more tests")

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := newModel.(*Model)

	if model.critiqueActive {
		t.Fatal("critiqueActive should be false after submitting critique")
	}
	if model.phase != PhasePRDGeneration {
		t.Fatalf("phase = %v, want PhasePRDGeneration", model.phase)
	}
	if !model.revisingPRD {
		t.Fatal("revisingPRD should be true after submitting critique")
	}
	if model.critiqueInput.Value() != "" {
		t.Fatalf("critiqueInput = %q, want cleared after submit", model.critiqueInput.Value())
	}
	if cmd == nil {
		t.Fatal("submitting critique should return a command")
	}
}

func TestUpdatePRDReviewCritiqueEnterAllowsEmptySubmission(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhasePRDReview
	m.prd = &prd.PRD{ProjectName: "P"}
	m.critiqueActive = true

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := newModel.(*Model)

	if model.phase != PhaseImplementation {
		t.Fatalf("phase = %v, want PhaseImplementation", model.phase)
	}
}

func TestUpdatePRDReviewCritiqueEscClearsDraftInput(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhasePRDReview
	m.prd = &prd.PRD{ProjectName: "P"}
	m.critiqueActive = true
	m.critiqueInput.SetValue("Discard this draft")

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := newModel.(*Model)

	if model.critiqueActive {
		t.Fatal("critiqueActive should be false after cancelling critique input")
	}
	if model.phase != PhasePRDReview {
		t.Fatalf("phase = %v, want PhasePRDReview", model.phase)
	}
	if model.critiqueInput.Value() != "" {
		t.Fatalf("critiqueInput = %q, want cleared draft input", model.critiqueInput.Value())
	}
}
