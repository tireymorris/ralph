package workflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/config"
	"ralph/internal/prd"
	"ralph/internal/runner"
)

func TestEventTypes(t *testing.T) {
	events := []Event{
		EventPRDGenerating{},
		EventPRDGenerated{PRD: &prd.PRD{}},
		EventPRDLoaded{PRD: &prd.PRD{}},
		EventStoryStarted{Story: &prd.Story{}, Iteration: 1},
		EventStoryCompleted{Story: &prd.Story{}, Success: true},
		EventOutput{Output{Text: "test", IsErr: false}},
		EventError{Err: nil},
		EventCompleted{},
		EventFailed{FailedStories: nil},
	}

	for _, e := range events {
		e.isEvent()
	}
}

func TestNewExecutor(t *testing.T) {
	cfg := config.DefaultConfig()
	ch := make(chan Event, 10)

	exec := NewExecutor(cfg, ch)
	if exec == nil {
		t.Fatal("NewExecutor() returned nil")
	}
	if exec.cfg != cfg {
		t.Error("cfg not set correctly")
	}
	if exec.eventsCh != ch {
		t.Error("eventsCh not set correctly")
	}
}

func TestEmitNilChannel(t *testing.T) {
	cfg := config.DefaultConfig()
	exec := NewExecutor(cfg, nil)

	exec.emit(EventCompleted{})
}

func TestEmitWithChannel(t *testing.T) {
	cfg := config.DefaultConfig()
	ch := make(chan Event, 10)
	exec := NewExecutor(cfg, ch)

	exec.emit(EventCompleted{})

	select {
	case e := <-ch:
		if _, ok := e.(EventCompleted); !ok {
			t.Errorf("received wrong event type: %T", e)
		}
	default:
		t.Error("no event received")
	}
}

func TestOutputStruct(t *testing.T) {
	o := Output{Text: "test", IsErr: true}

	if o.Text != "test" {
		t.Errorf("Text = %q, want %q", o.Text, "test")
	}
	if !o.IsErr {
		t.Error("IsErr = false, want true")
	}
}

func TestEventPRDGenerated(t *testing.T) {
	p := &prd.PRD{ProjectName: "Test"}
	e := EventPRDGenerated{PRD: p}

	if e.PRD.ProjectName != "Test" {
		t.Errorf("PRD.ProjectName = %q, want %q", e.PRD.ProjectName, "Test")
	}
}

func TestEventStoryStarted(t *testing.T) {
	s := &prd.Story{ID: "s1", Title: "Story 1"}
	e := EventStoryStarted{Story: s, Iteration: 5}

	if e.Story.ID != "s1" {
		t.Errorf("Story.ID = %q, want %q", e.Story.ID, "s1")
	}
	if e.Iteration != 5 {
		t.Errorf("Iteration = %d, want 5", e.Iteration)
	}
}

func TestEventStoryCompleted(t *testing.T) {
	s := &prd.Story{ID: "s1"}
	e := EventStoryCompleted{Story: s, Success: true}

	if !e.Success {
		t.Error("Success = false, want true")
	}
}

func TestEventFailed(t *testing.T) {
	stories := []*prd.Story{{ID: "s1"}, {ID: "s2"}}
	e := EventFailed{FailedStories: stories}

	if len(e.FailedStories) != 2 {
		t.Errorf("FailedStories length = %d, want 2", len(e.FailedStories))
	}
}

func TestForwardOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	eventsCh := make(chan Event, 10)
	exec := NewExecutor(cfg, eventsCh)

	outputCh := make(chan runner.OutputLine, 10)

	outputCh <- runner.OutputLine{Text: "test", IsErr: false}
	outputCh <- runner.OutputLine{Text: "error", IsErr: true}
	close(outputCh)

	exec.forwardOutput(outputCh)

	count := 0
	for range eventsCh {
		count++
		if count >= 2 {
			break
		}
	}

	if count != 2 {
		t.Errorf("forwarded %d events, want 2", count)
	}
}

func TestEventOutputEmbedding(t *testing.T) {
	e := EventOutput{Output{Text: "hello", IsErr: true}}
	if e.Text != "hello" {
		t.Errorf("Text = %q, want %q", e.Text, "hello")
	}
	if !e.IsErr {
		t.Error("IsErr = false, want true")
	}
}

func TestRunLoadSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"

	testPRD := &prd.PRD{ProjectName: "Test", Stories: []*prd.Story{{ID: "1", Title: "Story", Description: "Desc", AcceptanceCriteria: []string{"AC"}, Priority: 1}}}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("failed to save test PRD: %v", err)
	}

	ch := make(chan Event, 10)
	exec := NewExecutor(cfg, ch)

	p, err := exec.RunLoad(context.Background())
	if err != nil {
		t.Fatalf("RunLoad() error = %v", err)
	}
	if p.ProjectName != "Test" {
		t.Errorf("ProjectName = %q, want %q", p.ProjectName, "Test")
	}
}

func TestRunLoadError(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "nonexistent.json"

	ch := make(chan Event, 10)
	exec := NewExecutor(cfg, ch)

	_, err := exec.RunLoad(context.Background())
	if err == nil {
		t.Error("RunLoad() should return error when file doesn't exist")
	}
}

func TestRunImplementationAllCompleted(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories:     []*prd.Story{{ID: "1", Title: "Story", Description: "Desc", AcceptanceCriteria: []string{"AC"}, Priority: 1, Passes: true}},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("failed to save test PRD: %v", err)
	}

	ch := make(chan Event, 100)
	exec := NewExecutor(cfg, ch)

	err := exec.RunImplementation(context.Background(), testPRD)
	if err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}
}

func TestRunImplementationContextCancelled(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories:     []*prd.Story{{ID: "1", Title: "Story", Description: "Desc", AcceptanceCriteria: []string{"AC"}, Priority: 1, Passes: false}},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("failed to save test PRD: %v", err)
	}

	ch := make(chan Event, 100)
	exec := NewExecutor(cfg, ch)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := exec.RunImplementation(ctx, testPRD)
	if err == nil {
		t.Error("RunImplementation() should return error on context cancel")
	}
}

func TestRunImplementationNoNextStory(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"
	cfg.RetryAttempts = 1

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories:     []*prd.Story{{ID: "1", Title: "Story", Description: "Desc", AcceptanceCriteria: []string{"AC"}, Priority: 1, Passes: false, RetryCount: 5}},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("failed to save test PRD: %v", err)
	}

	ch := make(chan Event, 100)
	exec := NewExecutor(cfg, ch)

	err := exec.RunImplementation(context.Background(), testPRD)
	if err == nil {
		t.Error("RunImplementation() should return error when no pending stories")
	}
}

func TestRunImplementationMaxIterations(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"
	cfg.MaxIterations = 0

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories:     []*prd.Story{{ID: "1", Title: "Story", Description: "Desc", AcceptanceCriteria: []string{"AC"}, Priority: 1, Passes: false}},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("failed to save test PRD: %v", err)
	}

	ch := make(chan Event, 100)
	exec := NewExecutor(cfg, ch)

	err := exec.RunImplementation(context.Background(), testPRD)
	if err == nil {
		t.Error("RunImplementation() should return error on max iterations")
	}
}

func TestAllEventIsEventMethods(t *testing.T) {
	EventPRDGenerating{}.isEvent()
	EventPRDGenerated{}.isEvent()
	EventPRDLoaded{}.isEvent()
	EventStoryStarted{}.isEvent()
	EventStoryCompleted{}.isEvent()
	EventOutput{}.isEvent()
	EventError{}.isEvent()
	EventCompleted{}.isEvent()
	EventFailed{}.isEvent()
}

func TestNewExecutorWithRunner(t *testing.T) {
	cfg := config.DefaultConfig()
	ch := make(chan Event, 10)
	r := runner.New(cfg)

	exec := NewExecutorWithRunner(cfg, ch, r)
	if exec == nil {
		t.Fatal("NewExecutorWithRunner() returned nil")
	}
	if exec.runner != r {
		t.Error("runner not set correctly")
	}
}

func TestOutputVerboseField(t *testing.T) {
	o := Output{Text: "test", IsErr: false, Verbose: true}
	if !o.Verbose {
		t.Error("Verbose = false, want true")
	}
}

func TestPRDGetStory(t *testing.T) {
	p := &prd.PRD{
		Stories: []*prd.Story{
			{ID: "s1", Title: "Story 1"},
			{ID: "s2", Title: "Story 2"},
		},
	}

	s := p.GetStory("s1")
	if s == nil {
		t.Fatal("GetStory() returned nil")
	}
	if s.Title != "Story 1" {
		t.Errorf("Title = %q, want %q", s.Title, "Story 1")
	}

	s = p.GetStory("nonexistent")
	if s != nil {
		t.Error("GetStory() should return nil for nonexistent ID")
	}
}

func TestEmitChannelFull(t *testing.T) {
	cfg := config.DefaultConfig()
	ch := make(chan Event, 1)
	exec := NewExecutor(cfg, ch)

	ch <- EventCompleted{}

	exec.emit(EventCompleted{})
}

func setupTestPRDFile(t *testing.T, dir string, p *prd.PRD) *config.Config {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.WorkDir = dir
	cfg.PRDFile = "prd.json"

	if p != nil {
		prdPath := filepath.Join(dir, "prd.json")
		data := `{"project_name":"` + p.ProjectName + `","stories":[`
		for i, s := range p.Stories {
			if i > 0 {
				data += ","
			}
			passesStr := "false"
			if s.Passes {
				passesStr = "true"
			}
			data += `{"id":"` + s.ID + `","title":"` + s.Title + `","description":"` + s.Description + `","acceptance_criteria":["AC"],"priority":` + string(rune('0'+s.Priority)) + `,"passes":` + passesStr + `,"retry_count":` + string(rune('0'+s.RetryCount)) + `}`
		}
		data += `]}`
		if err := os.WriteFile(prdPath, []byte(data), 0644); err != nil {
			t.Fatalf("failed to write test PRD: %v", err)
		}
	}

	return cfg
}
