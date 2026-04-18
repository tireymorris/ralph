package tui

import (
	"testing"

	"ralph/internal/config"
	"ralph/internal/prd"
	"ralph/internal/workflow/events"
)

func TestHandleWorkflowEventPRDGenerating(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	cmd := m.handleWorkflowEvent(events.EventPRDGenerating{})
	if cmd != nil {
		t.Error("EventPRDGenerating should return nil cmd")
	}
	if m.phase != PhasePRDGeneration {
		t.Errorf("phase = %v, want PhasePRDGeneration", m.phase)
	}
}

func TestHandleWorkflowEventPRDGenerated(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	testPRD := &prd.PRD{ProjectName: "Test", Stories: []*prd.Story{{ID: "1"}}}
	m.handleWorkflowEvent(events.EventPRDGenerated{PRD: testPRD})

	if m.prd != testPRD {
		t.Error("prd should be set")
	}
}

func TestHandleWorkflowEventPRDLoaded(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	testPRD := &prd.PRD{ProjectName: "Test", Stories: []*prd.Story{{ID: "1", Passes: true}}}
	m.handleWorkflowEvent(events.EventPRDLoaded{PRD: testPRD})

	if m.prd != testPRD {
		t.Error("prd should be set")
	}
}

func TestHandleWorkflowEventStoryStarted(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	story := &prd.Story{ID: "1", Title: "Test Story"}
	m.handleWorkflowEvent(events.EventStoryStarted{Story: story, Iteration: 5})

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
	m.handleWorkflowEvent(events.EventStoryCompleted{Story: story, Success: true})

	if !m.prd.Stories[0].Passes {
		t.Error("story should be marked as passing")
	}
}

func TestHandleWorkflowEventStoryCompletedFailure(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	story := &prd.Story{ID: "1", Title: "Test", Passes: false}
	m.handleWorkflowEvent(events.EventStoryCompleted{Story: story, Success: false})

	if story.Passes {
		t.Error("story should remain not passing on failure")
	}
}

func TestHandleWorkflowEventCompleted(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	m.handleWorkflowEvent(events.EventCompleted{})

	if m.phase != PhaseCompleted {
		t.Errorf("phase = %v, want PhaseCompleted", m.phase)
	}
}

func TestHandleWorkflowEventFailed(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	m.handleWorkflowEvent(events.EventFailed{FailedStories: []*prd.Story{{ID: "1"}}})

	if m.phase != PhaseFailed {
		t.Errorf("phase = %v, want PhaseFailed", m.phase)
	}
}

func TestHandleWorkflowEventError(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	testErr := &testErrorType{msg: "error"}
	cmd := m.handleWorkflowEvent(events.EventError{Err: testErr})
	if cmd != nil {
		t.Error("EventError should return nil cmd")
	}
	if m.phase != PhaseFailed {
		t.Errorf("phase = %v, want PhaseFailed", m.phase)
	}
	if m.err != testErr {
		t.Error("err should be set")
	}
}

func TestHandleWorkflowEventOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	cmd := m.handleWorkflowEvent(events.EventOutput{Output: events.Output{Text: "test", IsErr: false}})
	if cmd != nil {
		t.Error("EventOutput should return nil cmd")
	}
}

func TestHandleWorkflowEventOutputVerboseFiltered(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false) // verbose=false

	// Verbose output on non-verbose model should not panic
	cmd := m.handleWorkflowEvent(events.EventOutput{Output: events.Output{Text: "verbose", IsErr: false, Verbose: true}})
	if cmd != nil {
		t.Error("EventOutput should return nil cmd")
	}
}

func TestHandleWorkflowEventOutputVerboseShown(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, true) // verbose=true

	// Verbose output on verbose model should not panic
	cmd := m.handleWorkflowEvent(events.EventOutput{Output: events.Output{Text: "verbose", IsErr: false, Verbose: true}})
	if cmd != nil {
		t.Error("EventOutput should return nil cmd")
	}
}

func TestUpdateWorkflowEventMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	newModel, cmd := m.Update(workflowEventMsg{event: events.EventCompleted{}})

	if model, ok := newModel.(*Model); ok {
		if model.phase != PhaseCompleted {
			t.Errorf("phase = %v, want PhaseCompleted", model.phase)
		}
	}
	if cmd == nil {
		t.Error("should return command to listen for more events")
	}
}

