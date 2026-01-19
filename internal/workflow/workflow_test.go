package workflow

import (
	"context"
	"errors"
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

type mockGenerator struct {
	prd *prd.PRD
	err error
}

func (m *mockGenerator) Generate(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) (*prd.PRD, error) {
	return m.prd, m.err
}

type mockImplementer struct {
	success bool
	err     error
}

func (m *mockImplementer) Implement(ctx context.Context, story *prd.Story, iteration int, p *prd.PRD, outputCh chan<- runner.OutputLine) (bool, error) {
	return m.success, m.err
}

type mockGitManager struct {
	err error
}

func (m *mockGitManager) IsRepository() bool {
	return true
}

func (m *mockGitManager) CurrentBranch() (string, error) {
	return "main", m.err
}

func (m *mockGitManager) BranchExists(name string) bool {
	return false
}

func (m *mockGitManager) CreateBranch(name string) error {
	return m.err
}

func (m *mockGitManager) Checkout(name string) error {
	return m.err
}

func (m *mockGitManager) HasChanges() bool {
	return true
}

func (m *mockGitManager) StageAll() error {
	return m.err
}

func (m *mockGitManager) Commit(message string) error {
	return m.err
}

func (m *mockGitManager) CommitStory(storyID, title, description string) error {
	return m.err
}

type mockStorage struct {
	prd     *prd.PRD
	loadErr error
	saveErr error
}

func (m *mockStorage) Load() (*prd.PRD, error) {
	return m.prd, m.loadErr
}

func (m *mockStorage) Save(p *prd.PRD) error {
	return m.saveErr
}

func (m *mockStorage) Delete() error {
	return nil
}

func TestNewExecutorWithDeps(t *testing.T) {
	cfg := config.DefaultConfig()
	ch := make(chan Event, 10)
	gen := &mockGenerator{}
	impl := &mockImplementer{}
	git := &mockGitManager{}
	storage := &mockStorage{}

	exec := NewExecutorWithDeps(cfg, ch, gen, impl, git, storage)
	if exec == nil {
		t.Fatal("NewExecutorWithDeps() returned nil")
	}
}

func TestRunGenerateSuccess(t *testing.T) {
	cfg := config.DefaultConfig()
	ch := make(chan Event, 10)
	testPRD := &prd.PRD{ProjectName: "Test"}
	gen := &mockGenerator{prd: testPRD}
	storage := &mockStorage{}

	exec := NewExecutorWithDeps(cfg, ch, gen, nil, nil, storage)

	p, err := exec.RunGenerate(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("RunGenerate() error = %v", err)
	}
	if p.ProjectName != "Test" {
		t.Errorf("ProjectName = %q, want %q", p.ProjectName, "Test")
	}
}

func TestRunGenerateGeneratorError(t *testing.T) {
	cfg := config.DefaultConfig()
	ch := make(chan Event, 10)
	gen := &mockGenerator{err: errors.New("gen error")}
	storage := &mockStorage{}

	exec := NewExecutorWithDeps(cfg, ch, gen, nil, nil, storage)

	_, err := exec.RunGenerate(context.Background(), "test")
	if err == nil {
		t.Error("RunGenerate() should return error")
	}
}

func TestRunGenerateSaveError(t *testing.T) {
	cfg := config.DefaultConfig()
	ch := make(chan Event, 10)
	gen := &mockGenerator{prd: &prd.PRD{}}
	storage := &mockStorage{saveErr: errors.New("save error")}

	exec := NewExecutorWithDeps(cfg, ch, gen, nil, nil, storage)

	_, err := exec.RunGenerate(context.Background(), "test")
	if err == nil {
		t.Error("RunGenerate() should return error on save failure")
	}
}

func TestRunLoadSuccess(t *testing.T) {
	cfg := config.DefaultConfig()
	ch := make(chan Event, 10)
	testPRD := &prd.PRD{ProjectName: "Loaded"}
	storage := &mockStorage{prd: testPRD}

	exec := NewExecutorWithDeps(cfg, ch, nil, nil, nil, storage)

	p, err := exec.RunLoad(context.Background())
	if err != nil {
		t.Fatalf("RunLoad() error = %v", err)
	}
	if p.ProjectName != "Loaded" {
		t.Errorf("ProjectName = %q, want %q", p.ProjectName, "Loaded")
	}
}

func TestRunLoadError(t *testing.T) {
	cfg := config.DefaultConfig()
	ch := make(chan Event, 10)
	storage := &mockStorage{loadErr: errors.New("load error")}

	exec := NewExecutorWithDeps(cfg, ch, nil, nil, nil, storage)

	_, err := exec.RunLoad(context.Background())
	if err == nil {
		t.Error("RunLoad() should return error")
	}
}

func TestRunImplementationAllCompleted(t *testing.T) {
	cfg := config.DefaultConfig()
	ch := make(chan Event, 10)
	storage := &mockStorage{}

	exec := NewExecutorWithDeps(cfg, ch, nil, nil, nil, storage)

	p := &prd.PRD{Stories: []*prd.Story{{ID: "1", Passes: true}}}
	err := exec.RunImplementation(context.Background(), p)
	if err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}
}

func TestRunImplementationWithBranch(t *testing.T) {
	cfg := config.DefaultConfig()
	ch := make(chan Event, 10)
	git := &mockGitManager{}
	storage := &mockStorage{}

	exec := NewExecutorWithDeps(cfg, ch, nil, nil, git, storage)

	p := &prd.PRD{BranchName: "feature/test", Stories: []*prd.Story{{ID: "1", Passes: true}}}
	err := exec.RunImplementation(context.Background(), p)
	if err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}
}

func TestRunImplementationBranchError(t *testing.T) {
	cfg := config.DefaultConfig()
	ch := make(chan Event, 10)
	git := &mockGitManager{err: errors.New("branch error")}
	storage := &mockStorage{}

	exec := NewExecutorWithDeps(cfg, ch, nil, nil, git, storage)

	p := &prd.PRD{BranchName: "feature/test", Stories: []*prd.Story{{ID: "1", Passes: true}}}
	err := exec.RunImplementation(context.Background(), p)
	if err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}
}

func TestRunImplementationContextCancelled(t *testing.T) {
	cfg := config.DefaultConfig()
	ch := make(chan Event, 10)
	impl := &mockImplementer{}
	storage := &mockStorage{}

	exec := NewExecutorWithDeps(cfg, ch, nil, impl, nil, storage)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := &prd.PRD{Stories: []*prd.Story{{ID: "1", Passes: false}}}
	err := exec.RunImplementation(ctx, p)
	if err == nil {
		t.Error("RunImplementation() should return error on context cancel")
	}
}

func TestRunImplementationNoNextStory(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RetryAttempts = 1
	ch := make(chan Event, 10)
	storage := &mockStorage{}

	exec := NewExecutorWithDeps(cfg, ch, nil, nil, nil, storage)

	p := &prd.PRD{Stories: []*prd.Story{{ID: "1", Passes: false, RetryCount: 5}}}
	err := exec.RunImplementation(context.Background(), p)
	if err == nil {
		t.Error("RunImplementation() should return error when no pending stories")
	}
}

func TestRunImplementationMaxIterations(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.MaxIterations = 0
	ch := make(chan Event, 10)
	storage := &mockStorage{}

	exec := NewExecutorWithDeps(cfg, ch, nil, nil, nil, storage)

	p := &prd.PRD{Stories: []*prd.Story{{ID: "1", Passes: false}}}
	err := exec.RunImplementation(context.Background(), p)
	if err == nil {
		t.Error("RunImplementation() should return error on max iterations")
	}
}

func TestRunImplementationStorySuccess(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RetryAttempts = 3
	ch := make(chan Event, 100)
	impl := &mockImplementer{success: true}
	storage := &mockStorage{}

	exec := NewExecutorWithDeps(cfg, ch, nil, impl, nil, storage)

	p := &prd.PRD{Stories: []*prd.Story{{ID: "1", Passes: false}}}
	err := exec.RunImplementation(context.Background(), p)
	if err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}
	if !p.Stories[0].Passes {
		t.Error("Story should be marked as passing")
	}
}

func TestRunImplementationStoryFailure(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RetryAttempts = 1
	ch := make(chan Event, 100)
	impl := &mockImplementer{success: false}
	storage := &mockStorage{}

	exec := NewExecutorWithDeps(cfg, ch, nil, impl, nil, storage)

	p := &prd.PRD{Stories: []*prd.Story{{ID: "1", Passes: false, RetryCount: 0}}}
	err := exec.RunImplementation(context.Background(), p)
	if err == nil {
		t.Error("RunImplementation() should return error when story fails")
	}
	if p.Stories[0].RetryCount != 1 {
		t.Errorf("RetryCount = %d, want 1", p.Stories[0].RetryCount)
	}
}

func TestRunImplementationStoryError(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RetryAttempts = 1
	ch := make(chan Event, 100)
	impl := &mockImplementer{err: errors.New("impl error")}
	storage := &mockStorage{}

	exec := NewExecutorWithDeps(cfg, ch, nil, impl, nil, storage)

	p := &prd.PRD{Stories: []*prd.Story{{ID: "1", Passes: false}}}
	err := exec.RunImplementation(context.Background(), p)
	if err == nil {
		t.Error("RunImplementation() should return error when impl errors")
	}
}

func TestRunImplementationSaveError(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RetryAttempts = 3
	ch := make(chan Event, 100)
	impl := &mockImplementer{success: true}
	storage := &mockStorage{saveErr: errors.New("save error")}

	exec := NewExecutorWithDeps(cfg, ch, nil, impl, nil, storage)

	p := &prd.PRD{Stories: []*prd.Story{{ID: "1", Passes: false}}}
	err := exec.RunImplementation(context.Background(), p)
	if err != nil {
		t.Fatalf("RunImplementation() should not return error on save warning, got %v", err)
	}
}

func TestDefaultPRDStorage(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.PRDFile = tmpDir + "/test.json"
	storage := &defaultPRDStorage{cfg: cfg}

	_, err := storage.Load()
	if err == nil {
		t.Error("Load() should error when file doesn't exist")
	}

	testPRD := &prd.PRD{ProjectName: "Test", Stories: []*prd.Story{{ID: "1"}}}
	err = storage.Save(testPRD)
	if err != nil {
		t.Errorf("Save() error = %v", err)
	}

	loaded, err := storage.Load()
	if err != nil {
		t.Errorf("Load() after Save error = %v", err)
	}
	if loaded.ProjectName != "Test" {
		t.Errorf("ProjectName = %q, want %q", loaded.ProjectName, "Test")
	}

	err = storage.Delete()
	if err != nil {
		t.Errorf("Delete() error = %v", err)
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
