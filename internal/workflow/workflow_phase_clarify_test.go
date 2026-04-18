package workflow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/config"
	"ralph/internal/prompt"
	"ralph/internal/runner"
)

// TestRunClarifyNoQuestionsFile verifies RunClarify returns nil,nil gracefully
// when the AI runner succeeds but writes no questions file.
func TestRunClarifyNoQuestionsFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		return nil // AI does not write a questions file
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	qas, err := exec.RunClarify(context.Background(), "add login")

	if err != nil {
		t.Fatalf("RunClarify() error = %v, want nil", err)
	}
	if qas != nil {
		t.Errorf("RunClarify() qas = %v, want nil when no questions file", qas)
	}
}

// TestRunClarifyRunnerError verifies RunClarify is non-fatal when the runner errors.
func TestRunClarifyRunnerError(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		return errors.New("runner failed")
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	qas, err := exec.RunClarify(context.Background(), "add login")

	if err != nil {
		t.Fatalf("RunClarify() should not propagate runner errors, got: %v", err)
	}
	if qas != nil {
		t.Errorf("RunClarify() qas should be nil on runner error")
	}
}

// TestRunClarifyInvalidJSON verifies RunClarify handles malformed questions file gracefully.
func TestRunClarifyInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		return os.WriteFile(filepath.Join(tmpDir, ClarifyingQuestionsFile), []byte("not json"), 0644)
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	qas, err := exec.RunClarify(context.Background(), "add login")

	if err != nil {
		t.Fatalf("RunClarify() should not error on bad JSON, got: %v", err)
	}
	if qas != nil {
		t.Errorf("RunClarify() qas should be nil on bad JSON")
	}
}

// TestRunClarifyEmptyArray verifies RunClarify returns nil when the AI writes [].
func TestRunClarifyEmptyArray(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		return os.WriteFile(filepath.Join(tmpDir, ClarifyingQuestionsFile), []byte("[]"), 0644)
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	qas, err := exec.RunClarify(context.Background(), "add login")

	if err != nil {
		t.Fatalf("RunClarify() error = %v", err)
	}
	if qas != nil {
		t.Errorf("RunClarify() should return nil for empty questions array")
	}
}

// TestRunClarifyWithQuestions verifies the full happy-path: AI writes questions,
// RunClarify emits EventClarifyingQuestions, and returns the answers sent to AnswersCh.
func TestRunClarifyWithQuestions(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		data := `["What language?", "Any auth requirements?"]`
		return os.WriteFile(filepath.Join(tmpDir, ClarifyingQuestionsFile), []byte(data), 0644)
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)

	// Consume the EventClarifyingQuestions from the channel and send answers back.
	expectedAnswers := []prompt.QuestionAnswer{
		{Question: "What language?", Answer: "Go"},
		{Question: "Any auth requirements?", Answer: "JWT"},
	}
	go func() {
		for event := range ch {
			if eq, ok := event.(EventClarifyingQuestions); ok {
				eq.AnswersCh <- expectedAnswers
				return
			}
		}
	}()

	qas, err := exec.RunClarify(context.Background(), "build an API")

	if err != nil {
		t.Fatalf("RunClarify() error = %v", err)
	}
	if len(qas) != 2 {
		t.Fatalf("RunClarify() returned %d answers, want 2", len(qas))
	}
	if qas[0].Answer != "Go" {
		t.Errorf("answer[0] = %q, want %q", qas[0].Answer, "Go")
	}
	if qas[1].Answer != "JWT" {
		t.Errorf("answer[1] = %q, want %q", qas[1].Answer, "JWT")
	}
}

// TestRunClarifyContextCancelled verifies RunClarify respects context cancellation
// while waiting for answers.
func TestRunClarifyContextCancelled(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		data := `["Blocking question?"]`
		return os.WriteFile(filepath.Join(tmpDir, ClarifyingQuestionsFile), []byte(data), 0644)
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately after RunClarify starts (once event is emitted).
	go func() {
		for event := range ch {
			if _, ok := event.(EventClarifyingQuestions); ok {
				cancel()
				return
			}
		}
	}()

	_, err := exec.RunClarify(ctx, "test")
	if err == nil {
		t.Error("RunClarify() should return error when context is cancelled")
	}
}

// TestRunClarifyNilChannel verifies RunClarify skips questions when no event channel.
func TestRunClarifyNilChannel(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir

	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		data := `["What language?"]`
		return os.WriteFile(filepath.Join(tmpDir, ClarifyingQuestionsFile), []byte(data), 0644)
	}

	// nil channel — simulates executor with no event consumer
	exec := NewExecutorWithRunner(cfg, nil, mock)
	qas, err := exec.RunClarify(context.Background(), "test")

	if err != nil {
		t.Fatalf("RunClarify() error = %v", err)
	}
	if qas != nil {
		t.Errorf("RunClarify() should return nil when no event channel")
	}
}

// TestRunClarifyQuestionsFileCleanedUp verifies the temporary questions file is
// deleted even when parsing succeeds (to avoid leaving state on disk).
func TestRunClarifyQuestionsFileCleanedUp(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		data := `["A question?"]`
		return os.WriteFile(filepath.Join(tmpDir, ClarifyingQuestionsFile), []byte(data), 0644)
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)

	go func() {
		for event := range ch {
			if eq, ok := event.(EventClarifyingQuestions); ok {
				eq.AnswersCh <- nil
				return
			}
		}
	}()

	exec.RunClarify(context.Background(), "test")

	if _, err := os.Stat(filepath.Join(tmpDir, ClarifyingQuestionsFile)); !os.IsNotExist(err) {
		t.Error("questions file should be deleted after RunClarify")
	}
}
