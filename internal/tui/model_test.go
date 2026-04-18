package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/config"
	"ralph/internal/prd"
	"ralph/internal/prompt"
	"ralph/internal/workflow"
)

func TestPhaseString(t *testing.T) {
	tests := []struct {
		phase Phase
		want  string
	}{
		{PhaseInit, "Initializing"},
		{PhaseClarifying, "Clarifying Questions"},
		{PhasePRDGeneration, "Phase 1: PRD Generation"},
		{PhasePRDReview, "PRD Review"},
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
	// PRD generation is now communicated via EventPRDGenerated workflow event
	newModel, _ := m.Update(workflowEventMsg{event: workflow.EventPRDGenerated{PRD: testPRD}})

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
	// After PRD generation, user is prompted to review before implementation
	newModel, _ := m.Update(workflowEventMsg{event: workflow.EventPRDGenerated{PRD: testPRD}})

	if model, ok := newModel.(*Model); ok {
		if model.phase != PhasePRDReview {
			t.Errorf("phase = %v, want PhasePRDReview", model.phase)
		}
	}
}

func TestUpdatePRDErrorMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	testErr := &testErrorType{msg: "test error"}
	// Errors are now communicated via EventError workflow event
	newModel, _ := m.Update(workflowEventMsg{event: workflow.EventError{Err: testErr}})

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

	cmd := m.handleWorkflowEvent(workflow.EventPRDGenerating{})
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

	if story.Passes {
		t.Error("story should remain not passing on failure")
	}
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

	testErr := &testErrorType{msg: "error"}
	cmd := m.handleWorkflowEvent(workflow.EventError{Err: testErr})
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

	cmd := m.handleWorkflowEvent(workflow.EventOutput{Output: workflow.Output{Text: "test", IsErr: false}})
	if cmd != nil {
		t.Error("EventOutput should return nil cmd")
	}
}

func TestHandleWorkflowEventOutputVerboseFiltered(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false) // verbose=false

	// Verbose output on non-verbose model should not panic
	cmd := m.handleWorkflowEvent(workflow.EventOutput{Output: workflow.Output{Text: "verbose", IsErr: false, Verbose: true}})
	if cmd != nil {
		t.Error("EventOutput should return nil cmd")
	}
}

func TestHandleWorkflowEventOutputVerboseShown(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, true) // verbose=true

	// Verbose output on verbose model should not panic
	cmd := m.handleWorkflowEvent(workflow.EventOutput{Output: workflow.Output{Text: "verbose", IsErr: false, Verbose: true}})
	if cmd != nil {
		t.Error("EventOutput should return nil cmd")
	}
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

func TestHandleWorkflowEventClarifyingQuestions(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	answersCh := make(chan []prompt.QuestionAnswer, 1)
	cmd := m.handleWorkflowEvent(workflow.EventClarifyingQuestions{
		Questions: []string{"Q1?", "Q2?"},
		AnswersCh: answersCh,
	})

	if cmd == nil {
		t.Fatal("EventClarifyingQuestions should return a command")
	}

	// Execute the returned command — it should produce a clarifyQuestionsMsg
	msg := cmd()
	cqm, ok := msg.(clarifyQuestionsMsg)
	if !ok {
		t.Fatalf("cmd() returned %T, want clarifyQuestionsMsg", msg)
	}
	if len(cqm.questions) != 2 {
		t.Errorf("questions count = %d, want 2", len(cqm.questions))
	}
	if cqm.questions[0] != "Q1?" {
		t.Errorf("questions[0] = %q, want %q", cqm.questions[0], "Q1?")
	}
	if cqm.answersCh == nil {
		t.Error("answersCh should not be nil")
	}
}

func TestUpdateClarifyQuestionsMsg(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	answersCh := make(chan []prompt.QuestionAnswer, 1)
	newModel, cmd := m.Update(clarifyQuestionsMsg{
		questions: []string{"Q1?", "Q2?"},
		answersCh: answersCh,
	})

	model, ok := newModel.(*Model)
	if !ok {
		t.Fatal("Update did not return *Model")
	}
	if model.phase != PhaseClarifying {
		t.Errorf("phase = %v, want PhaseClarifying", model.phase)
	}
	if len(model.clarifyQuestions) != 2 {
		t.Errorf("clarifyQuestions count = %d, want 2", len(model.clarifyQuestions))
	}
	if len(model.clarifyInputs) != 2 {
		t.Errorf("clarifyInputs count = %d, want 2", len(model.clarifyInputs))
	}
	if model.clarifyFocused != 0 {
		t.Errorf("clarifyFocused = %d, want 0", model.clarifyFocused)
	}
	if cmd == nil {
		t.Error("should return commands (textinput.Blink + ListenForEvents)")
	}
}

func TestUpdateClarifyingKeyEsc(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	// Put model in clarifying phase
	answersCh := make(chan []prompt.QuestionAnswer, 1)
	m.phase = PhaseClarifying
	m.clarifyQuestions = []string{"Q1?"}
	m.clarifyAnswersCh = answersCh
	ti := textinput.New()
	ti.Focus()
	m.clarifyInputs = []textinput.Model{ti}

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := newModel.(*Model)

	if model.phase != PhasePRDGeneration {
		t.Errorf("phase = %v, want PhasePRDGeneration after Esc", model.phase)
	}
	// Answers should have been sent (nil = skip)
	select {
	case answers := <-answersCh:
		if answers != nil {
			t.Errorf("Esc should send nil answers, got %v", answers)
		}
	default:
		t.Error("Esc should send answers to answersCh")
	}
}

func TestUpdateClarifyingKeyEnterNavigates(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	answersCh := make(chan []prompt.QuestionAnswer, 1)
	m.phase = PhaseClarifying
	m.clarifyQuestions = []string{"Q1?", "Q2?"}
	m.clarifyAnswersCh = answersCh
	m.clarifyFocused = 0
	ti1 := textinput.New()
	ti1.Focus()
	ti2 := textinput.New()
	m.clarifyInputs = []textinput.Model{ti1, ti2}

	// Enter on first field should move focus to second, not submit
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := newModel.(*Model)

	if model.phase != PhaseClarifying {
		t.Errorf("phase = %v, want PhaseClarifying (not yet on last field)", model.phase)
	}
	if model.clarifyFocused != 1 {
		t.Errorf("clarifyFocused = %d, want 1", model.clarifyFocused)
	}
	select {
	case <-answersCh:
		t.Error("Enter on non-last field should not submit answers")
	default:
		// correct — no submission yet
	}
}

func TestUpdateClarifyingKeyEnterSubmitsOnLast(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	answersCh := make(chan []prompt.QuestionAnswer, 1)
	m.phase = PhaseClarifying
	m.clarifyQuestions = []string{"Q1?"}
	m.clarifyAnswersCh = answersCh
	m.clarifyFocused = 0
	ti := textinput.New()
	ti.Focus()
	m.clarifyInputs = []textinput.Model{ti}

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := newModel.(*Model)

	if model.phase != PhasePRDGeneration {
		t.Errorf("phase = %v, want PhasePRDGeneration after submit", model.phase)
	}
	select {
	case answers := <-answersCh:
		if len(answers) != 1 {
			t.Errorf("got %d answers, want 1", len(answers))
		}
	default:
		t.Error("Enter on last field should submit answers")
	}
}

func TestBuildAnswers(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	// No questions
	if got := m.buildAnswers(); got != nil {
		t.Errorf("buildAnswers() with no questions = %v, want nil", got)
	}

	// With questions
	m.clarifyQuestions = []string{"Q1?", "Q2?"}
	ti1 := textinput.New()
	ti2 := textinput.New()
	m.clarifyInputs = []textinput.Model{ti1, ti2}

	got := m.buildAnswers()
	if len(got) != 2 {
		t.Fatalf("buildAnswers() = %d items, want 2", len(got))
	}
	if got[0].Question != "Q1?" {
		t.Errorf("got[0].Question = %q, want %q", got[0].Question, "Q1?")
	}
}

func TestSubmitClarifyingAnswersSendsAndTransitions(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	answersCh := make(chan []prompt.QuestionAnswer, 1)
	m.clarifyAnswersCh = answersCh
	m.phase = PhaseClarifying

	answers := []prompt.QuestionAnswer{{Question: "Q?", Answer: "A"}}
	cmds := m.submitClarifyingAnswers(answers)

	if m.phase != PhasePRDGeneration {
		t.Errorf("phase = %v, want PhasePRDGeneration", m.phase)
	}
	if m.clarifyAnswersCh != nil {
		t.Error("clarifyAnswersCh should be nil after submit")
	}
	if len(cmds) == 0 {
		t.Error("submitClarifyingAnswers should return ListenForEvents command")
	}
	select {
	case got := <-answersCh:
		if len(got) != 1 || got[0].Answer != "A" {
			t.Errorf("answersCh got %v, want answer 'A'", got)
		}
	default:
		t.Error("answers should have been sent to answersCh")
	}
}

func TestSubmitClarifyingAnswersNilChannelSafe(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	// clarifyAnswersCh is nil — should not panic
	m.phase = PhaseClarifying
	cmds := m.submitClarifyingAnswers(nil)
	if m.phase != PhasePRDGeneration {
		t.Errorf("phase = %v, want PhasePRDGeneration", m.phase)
	}
	if len(cmds) == 0 {
		t.Error("should still return ListenForEvents command even with nil channel")
	}
}

type testErrorType struct {
	msg string
}

func (e *testErrorType) Error() string {
	return e.msg
}
