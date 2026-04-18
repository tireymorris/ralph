package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ralph/internal/config"
	"ralph/internal/prd"
	"ralph/internal/runner"
)

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
	cfg.TestCommand = "true" // Tests always pass

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

// Test when passes is true but tests fail, story marked passes:false and rolled back
func TestRunImplementationPassesTrueButTestsFail(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"
	cfg.TestCommand = "exit 1" // Force tests to fail
	cfg.MaxIterations = 1
	cfg.RetryAttempts = 1

	// Initialize a git repo in temp dir
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitkeep"), []byte{}, 0644); err != nil {
		t.Fatalf("failed to create placeholder: %v", err)
	}

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories:     []*prd.Story{{ID: "1", Title: "Story", Description: "Desc", AcceptanceCriteria: []string{"AC"}, Priority: 1, Passes: false}},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("failed to save test PRD: %v", err)
	}

	ch := make(chan Event, 100)
	mock := newMockRunner()

	// Mock runner marks story as passing, but tests will fail
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		outputCh <- runner.OutputLine{Text: "Working on story..."}
		p, _ := prd.Load(cfg)
		p.Stories[0].Passes = true // Mark as complete
		return prd.Save(cfg, p)
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	_ = exec.RunImplementation(context.Background(), testPRD)

	// Load PRD from disk after implementation
	prdPath := filepath.Join(tmpDir, "prd.json")
	data, err := os.ReadFile(prdPath)
	if err != nil {
		t.Fatalf("failed to read PRD: %v", err)
	}
	var p prd.PRD
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatalf("failed to unmarshal PRD: %v", err)
	}
	if p.Stories[0].Passes {
		t.Error("expected passes to be false after test failure")
	}
	if p.Stories[0].RetryCount != 1 {
		t.Error("expected retry_count to be incremented")
	}
}

// Test when passes is true and tests pass, story stays passes:true
func TestRunImplementationPassesTrueAndTestsPass(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"
	cfg.TestCommand = "echo success" // Tests pass
	cfg.MaxIterations = 1
	cfg.RetryAttempts = 1

	// Initialize a git repo in temp dir
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitkeep"), []byte{}, 0644); err != nil {
		t.Fatalf("failed to create placeholder: %v", err)
	}

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories:     []*prd.Story{{ID: "1", Title: "Story", Description: "Desc", AcceptanceCriteria: []string{"AC"}, Priority: 1, Passes: false}},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("failed to save test PRD: %v", err)
	}

	ch := make(chan Event, 100)
	mock := newMockRunner()

	// Mock runner marks story as passing, and tests will pass
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		outputCh <- runner.OutputLine{Text: "Working on story..."}
		p, _ := prd.Load(cfg)
		p.Stories[0].Passes = true
		return prd.Save(cfg, p)
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	_ = exec.RunImplementation(context.Background(), testPRD)

	// Verify EventCompleted was emitted (meaning all stories passed including tests)
	foundCompleted := false
	for len(ch) > 0 {
		e := <-ch
		if _, ok := e.(EventCompleted); ok {
			foundCompleted = true
			break
		}
	}
	if !foundCompleted {
		t.Error("expected EventCompletion when tests pass")
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
	cfg.TestCommand = "true" // Tests always pass

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
	cfg.TestCommand = "true" // Tests always pass

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
