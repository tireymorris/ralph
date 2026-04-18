package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"ralph/internal/config"
	"ralph/internal/prd"
	"ralph/internal/prompt"
	"ralph/internal/workflow"
)

func TestNewHeadless(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "test prompt", true, false, false)

	if r == nil {
		t.Fatal("NewHeadless() returned nil")
	}
	if r.cfg != cfg {
		t.Error("cfg not set correctly")
	}
	if r.prompt != "test prompt" {
		t.Errorf("prompt = %q, want %q", r.prompt, "test prompt")
	}
	if !r.dryRun {
		t.Error("dryRun should be true")
	}
	if r.resume {
		t.Error("resume should be false")
	}
}

func TestNewHeadlessResume(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "", false, true, false)

	if !r.resume {
		t.Error("resume should be true")
	}
	if r.dryRun {
		t.Error("dryRun should be false")
	}
}

func TestNewHeadlessVerbose(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "test", false, false, true)

	if !r.verbose {
		t.Error("verbose should be true")
	}
}

func TestPrintStories(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "test", false, false, false)

	p := &prd.PRD{
		Stories: []*prd.Story{
			{Title: "Story 1", Priority: 1, Passes: true},
			{Title: "Story 2", Priority: 2, Passes: false},
		},
	}

	old := os.Stdout
	rr, w, _ := os.Pipe()
	os.Stdout = w

	r.printStories(p)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(rr)
	output := buf.String()

	if !strings.Contains(output, "Story 1") {
		t.Error("printStories() should contain story titles")
	}
	if !strings.Contains(output, "[x]") {
		t.Error("printStories() should show completed status")
	}
}

func TestHandleEventsPRDGenerating(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "test", false, false, false)

	eventsCh := make(chan workflow.Event, 10)
	doneCh := make(chan int, 1)

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	go r.handleEvents(eventsCh, doneCh)

	eventsCh <- workflow.EventPRDGenerating{}
	close(eventsCh)

	code := <-doneCh

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
}

func TestHandleEventsPRDGenerated(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "test", false, false, false)

	eventsCh := make(chan workflow.Event, 10)
	doneCh := make(chan int, 1)

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	go r.handleEvents(eventsCh, doneCh)

	eventsCh <- workflow.EventPRDGenerated{PRD: &prd.PRD{ProjectName: "Test", Stories: []*prd.Story{{ID: "1", Title: "S", Priority: 1}}}}
	close(eventsCh)

	<-doneCh

	w.Close()
	os.Stdout = old
}

func TestHandleEventsPRDLoaded(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "test", false, false, false)

	eventsCh := make(chan workflow.Event, 10)
	doneCh := make(chan int, 1)

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	go r.handleEvents(eventsCh, doneCh)

	eventsCh <- workflow.EventPRDLoaded{PRD: &prd.PRD{ProjectName: "Test", Stories: []*prd.Story{{ID: "1", Passes: true}}}}
	close(eventsCh)

	<-doneCh

	w.Close()
	os.Stdout = old
}

func TestHandleEventsStoryStarted(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RetryAttempts = 3
	r := NewHeadless(cfg, "test", false, false, false)

	eventsCh := make(chan workflow.Event, 10)
	doneCh := make(chan int, 1)

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	go r.handleEvents(eventsCh, doneCh)

	eventsCh <- workflow.EventStoryStarted{Story: &prd.Story{Title: "Test Story", RetryCount: 0}, Iteration: 1}
	close(eventsCh)

	<-doneCh

	w.Close()
	os.Stdout = old
}

func TestHandleEventsStoryCompletedSuccess(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "test", false, false, false)

	eventsCh := make(chan workflow.Event, 10)
	doneCh := make(chan int, 1)

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	go r.handleEvents(eventsCh, doneCh)

	eventsCh <- workflow.EventStoryCompleted{Story: &prd.Story{Title: "Test"}, Success: true}
	close(eventsCh)

	<-doneCh

	w.Close()
	os.Stdout = old
}

func TestHandleEventsStoryCompletedFailure(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "test", false, false, false)

	eventsCh := make(chan workflow.Event, 10)
	doneCh := make(chan int, 1)

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	go r.handleEvents(eventsCh, doneCh)

	eventsCh <- workflow.EventStoryCompleted{Story: &prd.Story{Title: "Test"}, Success: false}
	close(eventsCh)

	<-doneCh

	w.Close()
	os.Stdout = old
}

func TestHandleEventsOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "test", false, false, false)

	eventsCh := make(chan workflow.Event, 10)
	doneCh := make(chan int, 1)

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	go r.handleEvents(eventsCh, doneCh)

	eventsCh <- workflow.EventOutput{Output: workflow.Output{Text: "output text", IsErr: false}}
	eventsCh <- workflow.EventOutput{Output: workflow.Output{Text: "error text", IsErr: true}}
	close(eventsCh)

	<-doneCh

	w.Close()
	os.Stdout = old
}

func TestHandleEventsVerboseOutputFiltered(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "test", false, false, false)

	eventsCh := make(chan workflow.Event, 10)
	doneCh := make(chan int, 1)

	old := os.Stdout
	rr, w, _ := os.Pipe()
	os.Stdout = w

	go r.handleEvents(eventsCh, doneCh)

	eventsCh <- workflow.EventOutput{Output: workflow.Output{Text: "service bus log", IsErr: false, Verbose: true}}
	eventsCh <- workflow.EventOutput{Output: workflow.Output{Text: "tool call output", IsErr: false, Verbose: false}}
	close(eventsCh)

	<-doneCh

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(rr)
	output := buf.String()

	if strings.Contains(output, "service bus log") {
		t.Error("verbose output should be filtered when verbose=false")
	}
	if !strings.Contains(output, "tool call output") {
		t.Error("normal output should be shown")
	}
}

func TestHandleEventsVerboseOutputShown(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "test", false, false, true)

	eventsCh := make(chan workflow.Event, 10)
	doneCh := make(chan int, 1)

	old := os.Stdout
	rr, w, _ := os.Pipe()
	os.Stdout = w

	go r.handleEvents(eventsCh, doneCh)

	eventsCh <- workflow.EventOutput{Output: workflow.Output{Text: "service bus log", IsErr: false, Verbose: true}}
	close(eventsCh)

	<-doneCh

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(rr)
	output := buf.String()

	if !strings.Contains(output, "service bus log") {
		t.Error("verbose output should be shown when verbose=true")
	}
}

func TestHandleEventsError(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "test", false, false, false)

	eventsCh := make(chan workflow.Event, 10)
	doneCh := make(chan int, 1)

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	go r.handleEvents(eventsCh, doneCh)

	eventsCh <- workflow.EventError{Err: &testErr{msg: "test error"}}
	close(eventsCh)

	code := <-doneCh

	w.Close()
	os.Stdout = old

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestHandleEventsCompleted(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "test", false, false, false)

	eventsCh := make(chan workflow.Event, 10)
	doneCh := make(chan int, 1)

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	go r.handleEvents(eventsCh, doneCh)

	eventsCh <- workflow.EventCompleted{}
	close(eventsCh)

	code := <-doneCh

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
}

func TestHandleEventsFailed(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "test", false, false, false)

	eventsCh := make(chan workflow.Event, 10)
	doneCh := make(chan int, 1)

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	go r.handleEvents(eventsCh, doneCh)

	eventsCh <- workflow.EventFailed{FailedStories: []*prd.Story{{Title: "Failed", RetryCount: 3}}}
	close(eventsCh)

	code := <-doneCh

	w.Close()
	os.Stdout = old

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestHandleEventsFailedNoStories(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "test", false, false, false)

	eventsCh := make(chan workflow.Event, 10)
	doneCh := make(chan int, 1)

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	go r.handleEvents(eventsCh, doneCh)

	eventsCh <- workflow.EventFailed{FailedStories: []*prd.Story{}}
	close(eventsCh)

	code := <-doneCh

	w.Close()
	os.Stdout = old

	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

// TestHandleEventsClarifyingQuestions verifies that EventClarifyingQuestions
// causes the CLI to read stdin and send answers back on AnswersCh.
func TestHandleEventsClarifyingQuestions(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "test", false, false, false)

	// Pipe simulated stdin answers into the runner
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	origStdin := os.Stdin
	os.Stdin = stdinR
	defer func() { os.Stdin = origStdin }()

	// Write two answers then close (EOF)
	go func() {
		stdinW.WriteString("Go\n")
		stdinW.WriteString("JWT\n")
		stdinW.Close()
	}()

	eventsCh := make(chan workflow.Event, 10)
	doneCh := make(chan int, 1)

	// Capture stdout
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	go r.handleEvents(eventsCh, doneCh)

	answersCh := make(chan []prompt.QuestionAnswer, 1)
	eventsCh <- workflow.EventClarifyingQuestions{
		Questions: []string{"What language?", "Auth method?"},
		AnswersCh: answersCh,
	}
	close(eventsCh)

	<-doneCh
	w.Close()
	os.Stdout = old

	// Check answers were sent
	select {
	case answers := <-answersCh:
		if len(answers) != 2 {
			t.Fatalf("got %d answers, want 2", len(answers))
		}
		if answers[0].Answer != "Go" {
			t.Errorf("answer[0] = %q, want %q", answers[0].Answer, "Go")
		}
		if answers[1].Answer != "JWT" {
			t.Errorf("answer[1] = %q, want %q", answers[1].Answer, "JWT")
		}
	default:
		t.Error("no answers were sent to AnswersCh")
	}
}

// TestHandleEventsClarifyingQuestionsEmptyAnswers verifies blank stdin lines
// produce empty-string answers (not dropped).
func TestHandleEventsClarifyingQuestionsEmptyAnswers(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "test", false, false, false)

	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	origStdin := os.Stdin
	os.Stdin = stdinR
	defer func() { os.Stdin = origStdin }()

	go func() {
		stdinW.WriteString("\n") // blank answer
		stdinW.Close()
	}()

	eventsCh := make(chan workflow.Event, 10)
	doneCh := make(chan int, 1)

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	go r.handleEvents(eventsCh, doneCh)

	answersCh := make(chan []prompt.QuestionAnswer, 1)
	eventsCh <- workflow.EventClarifyingQuestions{
		Questions: []string{"Optional question?"},
		AnswersCh: answersCh,
	}
	close(eventsCh)

	<-doneCh
	w.Close()
	os.Stdout = old

	select {
	case answers := <-answersCh:
		if len(answers) != 1 {
			t.Fatalf("got %d answers, want 1", len(answers))
		}
		if answers[0].Answer != "" {
			t.Errorf("answer = %q, want empty string for blank input", answers[0].Answer)
		}
	default:
		t.Error("no answers were sent to AnswersCh")
	}
}

type testErr struct {
	msg string
}

func (e *testErr) Error() string {
	return e.msg
}

func TestHandleEventsPRDReview(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewHeadless(cfg, "test", false, false, false)

	eventsCh := make(chan workflow.Event, 10)
	doneCh := make(chan int, 1)

	go r.handleEvents(eventsCh, doneCh)

	testPRD := &prd.PRD{
		ProjectName: "Test Project",
		Stories: []*prd.Story{
			{ID: "story-1", Title: "Story 1", Priority: 1, Passes: false},
			{ID: "story-2", Title: "Story 2", Priority: 2, DependsOn: []string{"story-1"}, Passes: true},
		},
	}

	eventsCh <- workflow.EventPRDReview{PRD: testPRD}
	close(eventsCh)

	<-doneCh
}
