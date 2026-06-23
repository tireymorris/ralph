package tui

import (
	"strings"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/session"
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
	m.handleWorkflowEvent(events.EventStoryStarted{Story: story})

	if m.currentStory != story {
		t.Error("currentStory should be set")
	}
	if m.phase != PhaseImplementation {
		t.Errorf("phase = %v, want PhaseImplementation", m.phase)
	}
}

func TestHandleWorkflowEventStoryCompletedSuccess(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir

	story := &prd.Story{ID: "1", Title: "Test", Passes: true, Slices: []*prd.Slice{{ID: "slice-1", Behavior: "test", RedHint: "add failing test", Passes: true}}}
	if err := prd.Save(cfg, &prd.PRD{Stories: []*prd.Story{story}}); err != nil {
		t.Fatalf("Save PRD: %v", err)
	}

	m := NewModel(cfg, "test", false, false, false)
	stale := &prd.Story{ID: "1", Title: "Test", Passes: false}
	m.prd = &prd.PRD{Stories: []*prd.Story{stale}}
	m.handleWorkflowEvent(events.EventStoryCompleted{Story: story, Success: true})

	if !m.prd.Stories[0].Passes {
		t.Error("story should be marked as passing from disk reload")
	}
}

func TestHandleWorkflowEventStoryCompletedReloadsFromDisk(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir

	onDisk := &prd.PRD{
		ProjectName: "Disk",
		Stories: []*prd.Story{{
			ID:     "1",
			Title:  "Story One",
			Passes: true,
			Slices: []*prd.Slice{
				{ID: "slice-1", Behavior: "first", RedHint: "test", Passes: true},
				{ID: "slice-2", Behavior: "second", RedHint: "test", Passes: true},
			},
		}},
	}
	if err := prd.Save(cfg, onDisk); err != nil {
		t.Fatalf("Save PRD: %v", err)
	}

	m := NewModel(cfg, "test", false, false, false)
	story := &prd.Story{
		ID:     "1",
		Title:  "Story One",
		Passes: false,
		Slices: []*prd.Slice{
			{ID: "slice-1", Behavior: "first", RedHint: "test", Passes: false},
			{ID: "slice-2", Behavior: "second", RedHint: "test", Passes: false},
		},
	}
	m.prd = &prd.PRD{Stories: []*prd.Story{story}}
	m.currentStory = story
	m.phase = PhaseImplementation

	m.handleWorkflowEvent(events.EventStoryCompleted{Story: story, Success: true})

	if !m.prd.Stories[0].Passes {
		t.Fatal("story should reflect disk passes after reload")
	}
	if !m.prd.Stories[0].Slices[0].Passes || !m.prd.Stories[0].Slices[1].Passes {
		t.Fatal("all slices should reflect disk passes after reload")
	}
}

func TestHandleWorkflowEventSliceCompletedReloadsFromDisk(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir

	onDisk := &prd.PRD{
		ProjectName: "Disk",
		Stories: []*prd.Story{{
			ID:    "1",
			Title: "Story One",
			Slices: []*prd.Slice{
				{ID: "slice-1", Behavior: "first", RedHint: "test", Passes: true},
				{ID: "slice-2", Behavior: "second", RedHint: "test", Passes: false},
			},
		}},
	}
	if err := prd.Save(cfg, onDisk); err != nil {
		t.Fatalf("Save PRD: %v", err)
	}

	m := NewModel(cfg, "test", false, false, false)
	story := onDisk.Stories[0]
	stale := &prd.Story{
		ID:    "1",
		Title: "Story One",
		Slices: []*prd.Slice{
			{ID: "slice-1", Behavior: "first", RedHint: "test", Passes: false},
			{ID: "slice-2", Behavior: "second", RedHint: "test", Passes: false},
		},
	}
	m.prd = &prd.PRD{Stories: []*prd.Story{stale}}
	m.currentStory = stale
	m.phase = PhaseImplementation

	m.handleWorkflowEvent(events.EventSliceCompleted{StoryID: "1", SliceID: "slice-1"})

	if !m.prd.Stories[0].Slices[0].Passes {
		t.Fatal("completed slice should reflect disk after reload")
	}
	if m.snapshot.NextPendingSlice == nil || m.snapshot.NextPendingSlice.ID != "slice-2" {
		t.Fatalf("NextPendingSlice = %#v, want slice-2", m.snapshot.NextPendingSlice)
	}
	_ = story
}

func TestHandleWorkflowEventImplementationReviewStartedSetsPhaseAndActivity(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhaseCleanup
	m.currentStory = &prd.Story{ID: "1", Title: "Story One"}
	m.prd = &prd.PRD{Stories: []*prd.Story{m.currentStory}}

	m.handleWorkflowEvent(events.EventImplementationReviewStarted{Iteration: 2})

	if m.phase != PhaseCleanup {
		t.Fatalf("phase = %v, want PhaseCleanup", m.phase)
	}
	if m.activity.Kind != session.ActivityReview {
		t.Fatalf("activity.Kind = %q, want %q", m.activity.Kind, session.ActivityReview)
	}
	if m.activity.Iteration != 2 {
		t.Fatalf("activity.Iteration = %d, want 2", m.activity.Iteration)
	}
	if m.activity.StoryTitle != "Story One" {
		t.Fatalf("activity.StoryTitle = %q, want Story One", m.activity.StoryTitle)
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
	m.phase = PhaseFailed
	m.retryImplementation = true
	m.err = &testErrorType{msg: "stale"}

	m.handleWorkflowEvent(events.EventCompleted{})

	if m.phase != PhaseCompleted {
		t.Errorf("phase = %v, want PhaseCompleted", m.phase)
	}
	if m.retryImplementation {
		t.Error("retryImplementation should be cleared after completion")
	}
	if m.err != nil {
		t.Errorf("err = %v, want nil after completion", m.err)
	}
}

func TestHandleWorkflowEventError(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhasePRDGeneration

	testErr := &testErrorType{msg: "error"}
	cmd := m.handleWorkflowEvent(events.EventError{Err: testErr})
	if cmd != nil {
		t.Error("EventError should return nil cmd")
	}
	if m.err != testErr {
		t.Error("err should be set")
	}
	if m.phase != PhaseFailed {
		t.Errorf("phase = %v, want PhaseFailed", m.phase)
	}
	if m.retryImplementation {
		t.Error("retryImplementation should be false when not implementing")
	}
}

func TestHandleWorkflowEventErrorDuringImplementation(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhaseImplementation
	m.prd = &prd.PRD{ProjectName: "P"}

	testErr := &testErrorType{msg: "error"}
	m.handleWorkflowEvent(events.EventError{Err: testErr})
	if m.phase != PhaseFailed {
		t.Errorf("phase = %v, want PhaseFailed", m.phase)
	}
	if !m.retryImplementation {
		t.Error("retryImplementation should be true after implementation failure")
	}
}

func TestHandleWorkflowEventCleanupStarted(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhaseImplementation
	m.logger.SetSize(80, 10)

	cmd := m.handleWorkflowEvent(events.EventCleanupStarted{})
	if cmd != nil {
		t.Error("EventCleanupStarted should return nil cmd")
	}
	if m.phase != PhaseCleanup {
		t.Errorf("phase = %v, want PhaseCleanup", m.phase)
	}
	logView := m.logger.GetView().View()
	if !strings.Contains(logView, "Running post-implementation cleanup") {
		t.Errorf("log view should mention cleanup start, got %q", logView)
	}
}

func TestHandleWorkflowEventCleanupCompleted(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhaseCleanup
	m.logger.SetSize(80, 10)

	cmd := m.handleWorkflowEvent(events.EventCleanupCompleted{})
	if cmd != nil {
		t.Error("EventCleanupCompleted should return nil cmd")
	}
	if m.phase != PhaseCleanup {
		t.Errorf("phase = %v, want PhaseCleanup (EventCompleted handles the transition)", m.phase)
	}
	logView := m.logger.GetView().View()
	if !strings.Contains(logView, "Cleanup complete") {
		t.Errorf("log view should mention cleanup complete, got %q", logView)
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
	m := NewModel(cfg, "test", false, false, false)

	cmd := m.handleWorkflowEvent(events.EventOutput{Output: events.Output{Text: "verbose", IsErr: false, Verbose: true}})
	if cmd != nil {
		t.Error("EventOutput should return nil cmd")
	}
}

func TestHandleWorkflowEventOutputVerboseShown(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, true)

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
