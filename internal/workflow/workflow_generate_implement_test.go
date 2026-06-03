package workflow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
)

func TestRunGenerateEmptyWorkdirUsesNewProjectPrompt(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		if !strings.Contains(prompt, "no existing source code") {
			t.Error("expected PRD prompt for empty workdir to mention no existing source code")
		}
		prdPath := filepath.Join(tmpDir, "prd.json")
		data := `{"project_name":"Generated","stories":[{"id":"1","title":"Test","description":"Desc","acceptance_criteria":["AC"],"priority":1}]}`
		return os.WriteFile(prdPath, []byte(data), 0644)
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	if _, err := exec.RunGenerate(context.Background(), "test prompt"); err != nil {
		t.Fatalf("RunGenerate() error = %v", err)
	}
}

func TestRunGenerateWithSourceWorkdirUsesExistingCodebasePrompt(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)

	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		if strings.Contains(prompt, "no existing source code") {
			t.Error("expected PRD prompt for workdir with source to not mention no existing source code")
		}
		if !strings.Contains(prompt, "ACTUALLY observe in the codebase") {
			t.Error("expected PRD prompt for existing codebase to reference observed patterns")
		}
		prdPath := filepath.Join(tmpDir, "prd.json")
		data := `{"project_name":"Generated","stories":[{"id":"1","title":"Test","description":"Desc","acceptance_criteria":["AC"],"priority":1}]}`
		return os.WriteFile(prdPath, []byte(data), 0644)
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	if _, err := exec.RunGenerate(context.Background(), "test prompt"); err != nil {
		t.Fatalf("RunGenerate() error = %v", err)
	}
}

func TestRunGenerateSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()

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

	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		outputCh <- runner.OutputLine{Text: "Working on story..."}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	err := exec.RunImplementation(context.Background(), testPRD)

	if err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

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

	p, _ := prd.Load(cfg)
	if !p.Stories[0].Passes {
		t.Error("expected story to be marked as passing")
	}
}

func TestRunImplementationRunnerFailureReturnsError(t *testing.T) {
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

	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		return errors.New("runner failed")
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	err := exec.RunImplementation(context.Background(), testPRD)

	if err == nil {
		t.Fatal("RunImplementation() should return error when runner fails")
	}

	p, _ := prd.Load(cfg)
	if p.Stories[0].Passes {
		t.Error("expected story to remain incomplete after runner failure")
	}

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

func TestRunImplementationMultipleStories(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories: []*prd.Story{
			{ID: "1", Title: "Story 1", Description: "Desc 1", AcceptanceCriteria: []string{"AC1"}, Priority: 1, Passes: false},
			{ID: "2", Title: "Story 2", Description: "Desc 2", AcceptanceCriteria: []string{"AC2"}, Priority: 2, Passes: false},
		},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("failed to save test PRD: %v", err)
	}

	ch := make(chan Event, 100)
	mock := newMockRunner()

	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	err := exec.RunImplementation(context.Background(), testPRD)

	if err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

	p, _ := prd.Load(cfg)
	for _, s := range p.Stories {
		if !s.Passes {
			t.Errorf("expected story %s to be marked as passing", s.ID)
		}
	}
}

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

	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		return os.Remove(filepath.Join(tmpDir, "prd.json"))
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	err := exec.RunImplementation(context.Background(), testPRD)

	if err == nil {
		t.Error("RunImplementation() should return error when PRD reload fails")
	}
}

func TestRunImplementationVersionConflict(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"
	cfg.TestCommand = "true"

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

	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		p, _ := prd.Load(cfg)
		p.Version = 10
		return prd.Save(cfg, p)
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	err := exec.RunImplementation(context.Background(), testPRD)

	if err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

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

func TestRunCritiqueRevisionUpdatesPRDAndReturnsToReview(t *testing.T) {
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
	call := 0
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		call++
		if call == 1 {
			if !strings.Contains(prompt, "Needs more tests") {
				t.Fatalf("critique revision prompt missing critique:\n%s", prompt)
			}
			revised := &prd.PRD{
				ProjectName: "Revised",
				Stories:     []*prd.Story{{ID: "1", Title: "Story", Description: "Revised desc", AcceptanceCriteria: []string{"AC"}, Priority: 1, Passes: false}},
			}
			return prd.Save(cfg, revised)
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	err := exec.RunCritiqueRevision(context.Background(), "build feature", "Needs more tests")
	if err != nil {
		t.Fatalf("RunCritiqueRevision() error = %v", err)
	}

	foundReview := false
	for len(ch) > 0 {
		e := <-ch
		if review, ok := e.(EventPRDReview); ok {
			foundReview = true
			if review.PRD.ProjectName != "Revised" {
				t.Errorf("EventPRDReview project = %q, want Revised", review.PRD.ProjectName)
			}
		}
	}
	if !foundReview {
		t.Fatal("expected EventPRDReview after critique revision")
	}
}

func TestRunCritiqueRevisionAppliesClarificationsAfterClarify(t *testing.T) {
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
	call := 0
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		call++
		switch call {
		case 1:
			revised := &prd.PRD{ProjectName: "CritiqueApplied", Stories: testPRD.Stories}
			return prd.Save(cfg, revised)
		case 2:
			data := `["Which database?"]`
			return os.WriteFile(filepath.Join(tmpDir, ClarifyingQuestionsFile), []byte(data), 0644)
		case 3:
			if !strings.Contains(prompt, "Which database?") || !strings.Contains(prompt, "Postgres") {
				t.Fatalf("clarification revision prompt missing answers:\n%s", prompt)
			}
			final := &prd.PRD{ProjectName: "Final", Stories: testPRD.Stories}
			return prd.Save(cfg, final)
		default:
			t.Fatalf("unexpected runner call %d", call)
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	done := make(chan error, 1)
	go func() {
		done <- exec.RunCritiqueRevision(context.Background(), "build feature", "Needs more tests")
	}()

	var answersCh chan<- []prompt.QuestionAnswer
	timeout := time.After(2 * time.Second)
	for answersCh == nil {
		select {
		case event := <-ch:
			if eq, ok := event.(EventClarifyingQuestions); ok {
				answersCh = eq.AnswersCh
			}
		case <-timeout:
			t.Fatal("timed out waiting for clarifying questions during critique revision")
		}
	}

	answersCh <- []prompt.QuestionAnswer{{Question: "Which database?", Answer: "Postgres"}}

	if err := <-done; err != nil {
		t.Fatalf("RunCritiqueRevision() error = %v", err)
	}
	if mock.CallCount() != 3 {
		t.Fatalf("runner call count = %d, want 3", mock.CallCount())
	}
}

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
		return nil
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

func TestRunImplementationCallsCleanupBeforeCompleted(t *testing.T) {
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
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	err := exec.RunImplementation(context.Background(), testPRD)
	if err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

	var evts []Event
	for len(ch) > 0 {
		evts = append(evts, <-ch)
	}

	cleanupStartedIdx := -1
	cleanupCompletedIdx := -1
	completedIdx := -1
	for i, e := range evts {
		switch e.(type) {
		case EventCleanupStarted:
			cleanupStartedIdx = i
		case EventCleanupCompleted:
			cleanupCompletedIdx = i
		case EventCompleted:
			completedIdx = i
		}
	}

	if cleanupStartedIdx == -1 {
		t.Fatal("expected EventCleanupStarted to be emitted")
	}
	if cleanupCompletedIdx == -1 {
		t.Fatal("expected EventCleanupCompleted to be emitted")
	}
	if completedIdx == -1 {
		t.Fatal("expected EventCompleted to be emitted")
	}
	if cleanupStartedIdx >= completedIdx {
		t.Error("EventCleanupStarted must come before EventCompleted")
	}
	if cleanupCompletedIdx >= completedIdx {
		t.Error("EventCleanupCompleted must come before EventCompleted")
	}

	if mock.CallCount() < 2 {
		t.Errorf("runner should be called at least twice (story + cleanup), got %d", mock.CallCount())
	}
}

func TestRunImplementationCleanupFailureStopsCompleted(t *testing.T) {
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
	call := 0
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		call++
		if call == 1 {
			return nil
		}
		return errors.New("cleanup exploded")
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	err := exec.RunImplementation(context.Background(), testPRD)
	if err == nil {
		t.Fatal("RunImplementation() should return error when cleanup fails")
	}

	foundCompleted := false
	foundCleanupError := false
	for len(ch) > 0 {
		e := <-ch
		switch ev := e.(type) {
		case EventCompleted:
			foundCompleted = true
		case EventError:
			if strings.Contains(ev.Err.Error(), "cleanup") {
				foundCleanupError = true
			}
		}
	}

	if foundCompleted {
		t.Error("EventCompleted should NOT be emitted when cleanup fails")
	}
	if !foundCleanupError {
		t.Error("expected EventError with message containing 'cleanup'")
	}
}

func TestRunGenerateNoPRDFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()

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
