package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/config"
	"ralph/internal/prompt"
	"ralph/internal/workflow/events"
)

func TestHandleWorkflowEventClarifyingQuestions(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	answersCh := make(chan []prompt.QuestionAnswer, 1)
	cmd := m.handleWorkflowEvent(events.EventClarifyingQuestions{
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
