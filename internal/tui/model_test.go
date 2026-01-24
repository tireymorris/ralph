package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/config"
	"ralph/internal/prd"
	"ralph/internal/workflow"
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

func TestInit(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil")
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
	m := NewModel(cfg, "test", false, false, false)

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
	m := NewModel(cfg, "test", false, false, false)

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

func TestHandleWorkflowEventPRDGenerating(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	m.handleWorkflowEvent(workflow.EventPRDGenerating{})
}

func TestHandleWorkflowEventPRDGenerated(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	testPRD := &prd.PRD{ProjectName: "Test", Stories: []*prd.Story{{ID: "1"}}}
	m.handleWorkflowEvent(workflow.EventPRDGenerated{PRD: testPRD})

	if m.prd != testPRD {
		t.Error("prd should be set")
	}
}

func TestHandleWorkflowEventPRDLoaded(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	testPRD := &prd.PRD{ProjectName: "Test", Stories: []*prd.Story{{ID: "1", Passes: true}}}
	m.handleWorkflowEvent(workflow.EventPRDLoaded{PRD: testPRD})

	if m.prd != testPRD {
		t.Error("prd should be set")
	}
}

func TestHandleWorkflowEventStoryStarted(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	story := &prd.Story{ID: "1", Title: "Test Story"}
	m.handleWorkflowEvent(workflow.EventStoryStarted{Story: story, Iteration: 5})

	if m.currentStory != story {
		t.Error("currentStory should be set")
	}
	if m.iteration != 5 {
		t.Errorf("iteration = %d, want 5", m.iteration)
	}
}

func TestHandleWorkflowEventStoryCompletedSuccess(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	story := &prd.Story{ID: "1", Title: "Test", Passes: false}
	m.prd = &prd.PRD{Stories: []*prd.Story{story}}
	m.handleWorkflowEvent(workflow.EventStoryCompleted{Story: story, Success: true})

	if !m.prd.Stories[0].Passes {
		t.Error("story should be marked as passing")
	}
}

func TestHandleWorkflowEventStoryCompletedFailure(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	story := &prd.Story{ID: "1", Title: "Test", Passes: false}
	m.handleWorkflowEvent(workflow.EventStoryCompleted{Story: story, Success: false})
}

func TestHandleWorkflowEventCompleted(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	m.handleWorkflowEvent(workflow.EventCompleted{})

	if m.phase != PhaseCompleted {
		t.Errorf("phase = %v, want PhaseCompleted", m.phase)
	}
}

func TestHandleWorkflowEventFailed(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	m.handleWorkflowEvent(workflow.EventFailed{FailedStories: []*prd.Story{{ID: "1"}}})

	if m.phase != PhaseFailed {
		t.Errorf("phase = %v, want PhaseFailed", m.phase)
	}
}

func TestHandleWorkflowEventError(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	m.handleWorkflowEvent(workflow.EventError{Err: &testErrorType{msg: "error"}})
}

func TestHandleWorkflowEventOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	m.handleWorkflowEvent(workflow.EventOutput{Output: workflow.Output{Text: "test", IsErr: false}})
}

func TestHandleWorkflowEventOutputVerboseFiltered(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	m.handleWorkflowEvent(workflow.EventOutput{Output: workflow.Output{Text: "verbose", IsErr: false, Verbose: true}})
}

func TestHandleWorkflowEventOutputVerboseShown(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, true)

	m.handleWorkflowEvent(workflow.EventOutput{Output: workflow.Output{Text: "verbose", IsErr: false, Verbose: true}})
}

func TestUpdateWorkflowEventMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	newModel, cmd := m.Update(workflowEventMsg{event: workflow.EventCompleted{}})

	if model, ok := newModel.(*Model); ok {
		if model.phase != PhaseCompleted {
			t.Errorf("phase = %v, want PhaseCompleted", model.phase)
		}
	}
	if cmd == nil {
		t.Error("should return command to listen for more events")
	}
}

type testErrorType struct {
	msg string
}

func (e *testErrorType) Error() string {
	return e.msg
}
