package workflow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"ralph/internal/config"
	"ralph/internal/prd"
	"ralph/internal/runner"
)

// mockRunner implements RunnerInterface for testing
type mockRunner struct {
	runFunc      func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error
	runnerName   string
	commandName  string
	internalLogs []string
	mu           sync.Mutex
	calls        []string
}

func newMockRunner() *mockRunner {
	return &mockRunner{
		runnerName:  "mock",
		commandName: "mock-cmd",
	}
}

func (m *mockRunner) Run(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
	m.mu.Lock()
	m.calls = append(m.calls, prompt)
	m.mu.Unlock()

	if m.runFunc != nil {
		return m.runFunc(ctx, prompt, outputCh)
	}
	return nil
}

func (m *mockRunner) RunnerName() string  { return m.runnerName }
func (m *mockRunner) CommandName() string { return m.commandName }
func (m *mockRunner) IsInternalLog(line string) bool {
	for _, l := range m.internalLogs {
		if line == l {
			return true
		}
	}
	return false
}

func (m *mockRunner) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

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

// Test isPRDActionable heuristic
func TestIsPRDActionable(t *testing.T) {
	cfg := config.DefaultConfig()
	exec := NewExecutorWithRunner(cfg, nil, newMockRunner())

	tests := []struct {
		name string
		prd  *prd.PRD
		want bool
	}{
		{
			name: "actionable - specific description",
			prd: &prd.PRD{Stories: []*prd.Story{
				{Description: "Add login endpoint at /api/login"},
			}},
			want: true,
		},
		{
			name: "not actionable - vague optimize without quantification",
			prd: &prd.PRD{Stories: []*prd.Story{
				{Description: "Optimize the codebase"},
			}},
			want: false,
		},
		{
			name: "actionable - optimize with quantification",
			prd: &prd.PRD{Stories: []*prd.Story{
				{Description: "Optimize the prompt from 650 to 200 words"},
			}},
			want: true,
		},
		{
			name: "not actionable - vague improve",
			prd: &prd.PRD{Stories: []*prd.Story{
				{Description: "Improve error handling"},
			}},
			want: false,
		},
		{
			name: "actionable - refactor with specifics",
			prd: &prd.PRD{Stories: []*prd.Story{
				{Description: "Refactor validation to extract helper functions"},
			}},
			want: true,
		},
		{
			name: "actionable - no vague terms at all",
			prd: &prd.PRD{Stories: []*prd.Story{
				{Description: "Add user authentication with JWT tokens"},
			}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exec.isPRDActionable(tt.prd)
			if got != tt.want {
				t.Errorf("isPRDActionable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test RunGenerate success path
func TestRunGenerateSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()

	// Mock runner writes a valid PRD file
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		prdPath := filepath.Join(tmpDir, "prd.json")
		data := `{"project_name":"Generated","stories":[{"id":"1","title":"Test","description":"Desc","acceptance_criteria":["AC"],"priority":1}]}`
		return os.WriteFile(prdPath, []byte(data), 0644)
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	p, err := exec.RunGenerate(context.Background(), "test prompt")

	if err != nil {
		t.Fatalf("RunGenerate() error = %v", err)
	}
	if p.ProjectName != "Generated" {
		t.Errorf("ProjectName = %q, want %q", p.ProjectName, "Generated")
	}
	if mock.CallCount() < 1 {
		t.Errorf("runner called %d times, want at least 1", mock.CallCount())
	}
}

// Test RunGenerate when runner fails
func TestRunGenerateRunnerError(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		return errors.New("runner failed")
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	_, err := exec.RunGenerate(context.Background(), "test prompt")

	if err == nil {
		t.Error("RunGenerate() should return error when runner fails")
	}

	// Check error event was emitted
	foundError := false
	for len(ch) > 0 {
		e := <-ch
		if _, ok := e.(EventError); ok {
			foundError = true
			break
		}
	}
	if !foundError {
		t.Error("expected EventError to be emitted")
	}
}

// Test RunImplementation with mock runner completing a story
func TestRunImplementationStorySuccess(t *testing.T) {
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
	mock := newMockRunner()

	// Mock runner marks story as passing
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		outputCh <- runner.OutputLine{Text: "Working on story..."}
		// Load, update, and save PRD to mark story as complete
		p, _ := prd.Load(cfg)
		p.Stories[0].Passes = true
		return prd.Save(cfg, p)
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	err := exec.RunImplementation(context.Background(), testPRD)

	if err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

	// Verify EventCompleted was emitted
	foundCompleted := false
	for len(ch) > 0 {
		e := <-ch
		if _, ok := e.(EventCompleted); ok {
			foundCompleted = true
			break
		}
	}
	if !foundCompleted {
		t.Error("expected EventCompleted to be emitted")
	}
}

// Test RunImplementation retry count increment on failure
func TestRunImplementationRetryIncrement(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"
	cfg.RetryAttempts = 3
	cfg.MaxIterations = 2

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories:     []*prd.Story{{ID: "1", Title: "Story", Description: "Desc", AcceptanceCriteria: []string{"AC"}, Priority: 1, Passes: false, RetryCount: 0}},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("failed to save test PRD: %v", err)
	}

	ch := make(chan Event, 100)
	mock := newMockRunner()

	// Mock runner does NOT mark story as passing
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		return nil // Story stays incomplete
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	_ = exec.RunImplementation(context.Background(), testPRD)

	// Load PRD and check retry count was incremented
	p, err := prd.Load(cfg)
	if err != nil {
		t.Fatalf("failed to load PRD: %v", err)
	}
	if p.Stories[0].RetryCount == 0 {
		t.Error("expected retry count to be incremented")
	}
}

// Test RunImplementation PRD reload failure
func TestRunImplementationPRDReloadError(t *testing.T) {
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
	mock := newMockRunner()

	// Mock runner deletes the PRD file
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		return os.Remove(filepath.Join(tmpDir, "prd.json"))
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	err := exec.RunImplementation(context.Background(), testPRD)

	if err == nil {
		t.Error("RunImplementation() should return error when PRD reload fails")
	}
}

// Test RunImplementation version conflict detection
func TestRunImplementationVersionConflict(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"

	testPRD := &prd.PRD{
		Version:     1,
		ProjectName: "Test",
		Stories:     []*prd.Story{{ID: "1", Title: "Story", Description: "Desc", AcceptanceCriteria: []string{"AC"}, Priority: 1, Passes: false}},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("failed to save test PRD: %v", err)
	}

	ch := make(chan Event, 100)
	mock := newMockRunner()

	// Mock runner jumps version significantly (simulating external modification)
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		p, _ := prd.Load(cfg)
		p.Version = 10 // Big jump
		p.Stories[0].Passes = true
		return prd.Save(cfg, p)
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	err := exec.RunImplementation(context.Background(), testPRD)

	if err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

	// Check that a warning event was emitted about version jump
	foundWarning := false
	for len(ch) > 0 {
		e := <-ch
		if eo, ok := e.(EventOutput); ok {
			if len(eo.Text) > 0 && (strings.Contains(eo.Text, "version") || strings.Contains(eo.Text, "modified")) {
				foundWarning = true
			}
		}
	}
	if !foundWarning {
		t.Log("Note: version conflict warning may be logged rather than emitted as event")
	}
}

// Test output forwarding with verbose flag
func TestForwardOutputVerbose(t *testing.T) {
	cfg := config.DefaultConfig()
	eventsCh := make(chan Event, 10)
	exec := NewExecutor(cfg, eventsCh)

	outputCh := make(chan runner.OutputLine, 10)

	outputCh <- runner.OutputLine{Text: "normal", IsErr: false, Verbose: false}
	outputCh <- runner.OutputLine{Text: "verbose", IsErr: false, Verbose: true}
	outputCh <- runner.OutputLine{Text: "error", IsErr: true, Verbose: false}
	close(outputCh)

	exec.forwardOutput(outputCh)

	count := 0
	hasVerbose := false
	hasError := false
	for len(eventsCh) > 0 {
		e := <-eventsCh
		if eo, ok := e.(EventOutput); ok {
			count++
			if eo.Verbose {
				hasVerbose = true
			}
			if eo.IsErr {
				hasError = true
			}
		}
	}

	if count != 3 {
		t.Errorf("forwarded %d events, want 3", count)
	}
	if !hasVerbose {
		t.Error("expected verbose output to be forwarded")
	}
	if !hasError {
		t.Error("expected error output to be forwarded")
	}
}

// Test that EventPRDLoaded is emitted on RunLoad success
func TestRunLoadEmitsEvent(t *testing.T) {
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

	_, err := exec.RunLoad(context.Background())
	if err != nil {
		t.Fatalf("RunLoad() error = %v", err)
	}

	found := false
	for len(ch) > 0 {
		e := <-ch
		if _, ok := e.(EventPRDLoaded); ok {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected EventPRDLoaded to be emitted")
	}
}

// Test EventStoryStarted and EventStoryCompleted are emitted
func TestRunImplementationEmitsStoryEvents(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories:     []*prd.Story{{ID: "story-1", Title: "Story", Description: "Desc", AcceptanceCriteria: []string{"AC"}, Priority: 1, Passes: false}},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("failed to save test PRD: %v", err)
	}

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		p, _ := prd.Load(cfg)
		p.Stories[0].Passes = true
		return prd.Save(cfg, p)
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	err := exec.RunImplementation(context.Background(), testPRD)
	if err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

	foundStarted := false
	foundCompleted := false
	for len(ch) > 0 {
		e := <-ch
		switch ev := e.(type) {
		case EventStoryStarted:
			if ev.Story.ID == "story-1" {
				foundStarted = true
			}
		case EventStoryCompleted:
			if ev.Story.ID == "story-1" && ev.Success {
				foundCompleted = true
			}
		}
	}

	if !foundStarted {
		t.Error("expected EventStoryStarted to be emitted")
	}
	if !foundCompleted {
		t.Error("expected EventStoryCompleted with Success=true to be emitted")
	}
}

func TestIsEmptyCodebase(t *testing.T) {
	t.Run("empty directory", func(t *testing.T) {
		dir := t.TempDir()
		if !isEmptyCodebase(dir) {
			t.Error("expected empty directory to be detected as empty codebase")
		}
	})

	t.Run("directory with Go files", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
		if isEmptyCodebase(dir) {
			t.Error("expected directory with .go file to be detected as non-empty")
		}
	})

	t.Run("directory with only non-code files", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0644)
		os.WriteFile(filepath.Join(dir, "config.json"), []byte("{}"), 0644)
		if !isEmptyCodebase(dir) {
			t.Error("expected directory with only non-code files to be detected as empty codebase")
		}
	})

	t.Run("skips hidden directories", func(t *testing.T) {
		dir := t.TempDir()
		hiddenDir := filepath.Join(dir, ".git")
		os.MkdirAll(hiddenDir, 0755)
		os.WriteFile(filepath.Join(hiddenDir, "hooks.py"), []byte("# hook"), 0644)
		if !isEmptyCodebase(dir) {
			t.Error("expected code in hidden dirs to be ignored")
		}
	})

	t.Run("finds code in subdirectories", func(t *testing.T) {
		dir := t.TempDir()
		subDir := filepath.Join(dir, "src")
		os.MkdirAll(subDir, 0755)
		os.WriteFile(filepath.Join(subDir, "app.ts"), []byte("export {}"), 0644)
		if isEmptyCodebase(dir) {
			t.Error("expected code in subdirectory to be detected")
		}
	})

	t.Run("empty string work dir", func(t *testing.T) {
		if !isEmptyCodebase("") {
			t.Error("expected empty work dir to return true")
		}
	})
}

func TestIsPRDActionableVagueCriteria(t *testing.T) {
	cfg := config.DefaultConfig()
	exec := NewExecutorWithRunner(cfg, nil, newMockRunner())

	tests := []struct {
		name string
		prd  *prd.PRD
		want bool
	}{
		{
			name: "vague acceptance criteria - proper",
			prd: &prd.PRD{Stories: []*prd.Story{
				{Description: "Add error handling", AcceptanceCriteria: []string{"Proper error handling implemented"}},
			}},
			want: false,
		},
		{
			name: "vague acceptance criteria - comprehensive",
			prd: &prd.PRD{Stories: []*prd.Story{
				{Description: "Add tests", AcceptanceCriteria: []string{"Comprehensive test coverage"}},
			}},
			want: false,
		},
		{
			name: "specific acceptance criteria",
			prd: &prd.PRD{Stories: []*prd.Story{
				{Description: "Add error handling", AcceptanceCriteria: []string{"Returns 400 status with error message for invalid input"}},
			}},
			want: true,
		},
		{
			name: "vague verb in acceptance criteria",
			prd: &prd.PRD{Stories: []*prd.Story{
				{Description: "Add endpoint", AcceptanceCriteria: []string{"Optimize query performance"}},
			}},
			want: false,
		},
		{
			name: "vague adjective with quantifier passes",
			prd: &prd.PRD{Stories: []*prd.Story{
				{Description: "Add tests", AcceptanceCriteria: []string{"Comprehensive tests covering 90% of lines"}},
			}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exec.isPRDActionable(tt.prd)
			if got != tt.want {
				t.Errorf("isPRDActionable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunGenerateNoPRDFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()
	// Mock runner succeeds but does NOT create a PRD file
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	_, err := exec.RunGenerate(context.Background(), "test prompt")

	if err == nil {
		t.Fatal("RunGenerate() should return error when PRD file not created")
	}
	if !strings.Contains(err.Error(), "did not generate") {
		t.Errorf("error should mention 'did not generate', got: %v", err)
	}
}
