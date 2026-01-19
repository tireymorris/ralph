package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/config"
	"ralph/internal/prd"
)

func TestPhaseString(t *testing.T) {
	tests := []struct {
		phase Phase
		want  string
	}{
		{PhaseInit, "Initializing"},
		{PhasePRDGeneration, "Phase 1: PRD Generation"},
		{PhaseImplementation, "Phase 2: Implementation"},
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

func TestNewModel(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test prompt", true, false)

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
	if m.ctx == nil {
		t.Error("ctx should not be nil")
	}
	if m.cancelFunc == nil {
		t.Error("cancelFunc should not be nil")
	}
	if m.outputCh == nil {
		t.Error("outputCh should not be nil")
	}
	if m.maxLogs != 100 {
		t.Errorf("maxLogs = %d, want 100", m.maxLogs)
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
			name:     "failed no prd",
			phase:    PhaseFailed,
			prd:      nil,
			wantCode: 1,
		},
		{
			name:  "failed with some completed",
			phase: PhaseFailed,
			prd: &prd.PRD{
				Stories: []*prd.Story{
					{Passes: true},
					{Passes: false},
				},
			},
			wantCode: 2,
		},
		{
			name:  "failed with none completed",
			phase: PhaseFailed,
			prd: &prd.PRD{
				Stories: []*prd.Story{
					{Passes: false},
				},
			},
			wantCode: 1,
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

func TestAddLog(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false)

	m.addLog("line 1")
	m.addLog("line 2")

	if len(m.logs) != 2 {
		t.Errorf("logs length = %d, want 2", len(m.logs))
	}
	if m.logs[0] != "line 1" {
		t.Errorf("logs[0] = %q, want %q", m.logs[0], "line 1")
	}
	if m.logs[1] != "line 2" {
		t.Errorf("logs[1] = %q, want %q", m.logs[1], "line 2")
	}
}

func TestAddLogMaxLogs(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false)
	m.maxLogs = 3

	m.addLog("1")
	m.addLog("2")
	m.addLog("3")
	m.addLog("4")

	if len(m.logs) != 3 {
		t.Errorf("logs length = %d, want 3", len(m.logs))
	}
	if m.logs[0] != "2" {
		t.Errorf("logs[0] = %q, want %q", m.logs[0], "2")
	}
}

func TestInit(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false)

	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil")
	}
}

func TestUpdateKeyMsgQuit(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false)

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
	m := NewModel(cfg, "test", false, false)

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if model, ok := newModel.(*Model); ok {
		if !model.quitting {
			t.Error("quitting should be true after Ctrl+C")
		}
	}
}

func TestUpdateWindowSizeMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false)

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

func TestUpdateOutputMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false)

	newModel, _ := m.Update(outputMsg{Text: "test output"})

	if model, ok := newModel.(*Model); ok {
		if len(model.logs) == 0 || model.logs[len(model.logs)-1] != "test output" {
			t.Error("output should be added to logs")
		}
	}
}

func TestUpdatePRDGeneratedMsgDryRun(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", true, false)

	testPRD := &prd.PRD{ProjectName: "Test", Stories: []*prd.Story{{ID: "1"}}}
	newModel, _ := m.Update(prdGeneratedMsg{prd: testPRD})

	if model, ok := newModel.(*Model); ok {
		if model.prd != testPRD {
			t.Error("prd should be set")
		}
		if model.phase != PhaseCompleted {
			t.Errorf("phase = %v, want PhaseCompleted for dry run", model.phase)
		}
	}
}

func TestUpdatePRDGeneratedMsgImplement(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false)

	testPRD := &prd.PRD{ProjectName: "Test", Stories: []*prd.Story{{ID: "1"}}}
	newModel, _ := m.Update(prdGeneratedMsg{prd: testPRD})

	if model, ok := newModel.(*Model); ok {
		if model.phase != PhaseImplementation {
			t.Errorf("phase = %v, want PhaseImplementation", model.phase)
		}
	}
}

func TestUpdatePRDErrorMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false)

	testErr := &testErrorType{msg: "test error"}
	newModel, _ := m.Update(prdErrorMsg{err: testErr})

	if model, ok := newModel.(*Model); ok {
		if model.err != testErr {
			t.Error("err should be set")
		}
		if model.phase != PhaseFailed {
			t.Errorf("phase = %v, want PhaseFailed", model.phase)
		}
	}
}

func TestUpdateStoryStartMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false)
	m.iteration = 0

	story := &prd.Story{ID: "1", Title: "Test Story"}
	newModel, _ := m.Update(storyStartMsg{story: story})

	if model, ok := newModel.(*Model); ok {
		if model.currentStory != story {
			t.Error("currentStory should be set")
		}
		if model.iteration != 1 {
			t.Errorf("iteration = %d, want 1", model.iteration)
		}
	}
}

func TestUpdateStoryCompleteMsgSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.PRDFile = tmpDir + "/prd.json"

	m := NewModel(cfg, "test", false, false)
	story := &prd.Story{ID: "1", Title: "Test", Passes: false}
	m.currentStory = story
	m.prd = &prd.PRD{Stories: []*prd.Story{story}}

	newModel, _ := m.Update(storyCompleteMsg{success: true})

	if model, ok := newModel.(*Model); ok {
		if !model.currentStory.Passes {
			t.Error("story should be marked as passing")
		}
	}
}

func TestUpdateStoryCompleteMsgSaveError(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.PRDFile = "/nonexistent/dir/prd.json"

	m := NewModel(cfg, "test", false, false)
	story := &prd.Story{ID: "1", Title: "Test", Passes: false}
	m.currentStory = story
	m.prd = &prd.PRD{Stories: []*prd.Story{story}}

	m.Update(storyCompleteMsg{success: true})
}

func TestUpdateStoryCompleteMsgFailure(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.PRDFile = tmpDir + "/prd.json"

	m := NewModel(cfg, "test", false, false)
	story := &prd.Story{ID: "1", Title: "Test", Passes: false, RetryCount: 0}
	m.currentStory = story
	m.prd = &prd.PRD{Stories: []*prd.Story{story}}

	newModel, _ := m.Update(storyCompleteMsg{success: false})

	if model, ok := newModel.(*Model); ok {
		if model.currentStory.RetryCount != 1 {
			t.Errorf("retry count = %d, want 1", model.currentStory.RetryCount)
		}
	}
}

func TestUpdateStoryErrorMsg(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.PRDFile = tmpDir + "/prd.json"

	m := NewModel(cfg, "test", false, false)
	story := &prd.Story{ID: "1", Title: "Test", Passes: false, RetryCount: 0}
	m.currentStory = story
	m.prd = &prd.PRD{Stories: []*prd.Story{story}}

	newModel, _ := m.Update(storyErrorMsg{err: &testErrorType{msg: "error"}})

	if model, ok := newModel.(*Model); ok {
		if model.currentStory.RetryCount != 1 {
			t.Errorf("retry count = %d, want 1", model.currentStory.RetryCount)
		}
	}
}

func TestUpdatePhaseChangeMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false)

	newModel, _ := m.Update(phaseChangeMsg(PhaseCompleted))

	if model, ok := newModel.(*Model); ok {
		if model.phase != PhaseCompleted {
			t.Errorf("phase = %v, want PhaseCompleted", model.phase)
		}
	}
}

func TestUpdateSpinnerTickMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false)

	_, cmd := m.Update(m.spinner.Tick())
	if cmd == nil {
		t.Error("spinner tick should return a command")
	}
}

type testErrorType struct {
	msg string
}

func (e *testErrorType) Error() string {
	return e.msg
}
