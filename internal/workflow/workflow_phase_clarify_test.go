package workflow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/runner"
)

func TestRunClarifyEmptyWorkdirUsesNewProjectPrompt(t *testing.T) {
	tmpDir := t.TempDir()
	ch := make(chan Event, 100)
	exec := newClarifyExecutor(t, tmpDir, func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if !strings.Contains(p, "new project (no existing source code)") {
			t.Error("expected clarify prompt for empty workdir to mention new project")
		}
		return nil
	}, ch)
	if _, err := exec.RunClarify(context.Background(), "add login"); err != nil {
		t.Fatalf("RunClarify() error = %v", err)
	}
}

func TestRunClarifyWithSourceWorkdirUsesExistingCodebasePrompt(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)

	ch := make(chan Event, 100)
	exec := newClarifyExecutor(t, tmpDir, func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if strings.Contains(p, "new project (no existing source code)") {
			t.Error("expected clarify prompt for workdir with source to not mention new project")
		}
		if !strings.Contains(p, "an existing codebase") {
			t.Error("expected clarify prompt for workdir with source to mention existing codebase")
		}
		return nil
	}, ch)
	if _, err := exec.RunClarify(context.Background(), "add login"); err != nil {
		t.Fatalf("RunClarify() error = %v", err)
	}
}

func TestRunClarifyNoQuestionsFile(t *testing.T) {
	tmpDir := t.TempDir()
	ch := make(chan Event, 100)
	exec := newClarifyExecutor(t, tmpDir, func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		return nil
	}, ch)
	qas, err := exec.RunClarify(context.Background(), "add login")

	if err != nil {
		t.Fatalf("RunClarify() error = %v, want nil", err)
	}
	if qas != nil {
		t.Errorf("RunClarify() qas = %v, want nil when no questions file", qas)
	}
}

func TestRunClarifyRunnerError(t *testing.T) {
	tmpDir := t.TempDir()
	ch := make(chan Event, 100)
	exec := newClarifyExecutor(t, tmpDir, func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		return errors.New("runner failed")
	}, ch)
	qas, err := exec.RunClarify(context.Background(), "add login")

	if err != nil {
		t.Fatalf("RunClarify() should not propagate runner errors, got: %v", err)
	}
	if qas != nil {
		t.Errorf("RunClarify() qas should be nil on runner error")
	}
}

func TestRunClarifyInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	ch := make(chan Event, 100)
	exec := newClarifyExecutor(t, tmpDir, func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		return writeQuestionsFile(t, tmpDir, "not json")
	}, ch)
	qas, err := exec.RunClarify(context.Background(), "add login")

	if err != nil {
		t.Fatalf("RunClarify() should not error on bad JSON, got: %v", err)
	}
	if qas != nil {
		t.Errorf("RunClarify() qas should be nil on bad JSON")
	}
}

func TestRunClarifyEmptyArray(t *testing.T) {
	tmpDir := t.TempDir()
	ch := make(chan Event, 100)
	exec := newClarifyExecutor(t, tmpDir, func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		return writeQuestionsFile(t, tmpDir, "[]")
	}, ch)
	qas, err := exec.RunClarify(context.Background(), "add login")

	if err != nil {
		t.Fatalf("RunClarify() error = %v", err)
	}
	if qas != nil {
		t.Errorf("RunClarify() should return nil for empty questions array")
	}
}

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

func TestRunClarifyNilChannel(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir

	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		data := `["What language?"]`
		return os.WriteFile(filepath.Join(tmpDir, ClarifyingQuestionsFile), []byte(data), 0644)
	}

	exec := NewExecutorWithRunner(cfg, nil, mock)
	qas, err := exec.RunClarify(context.Background(), "test")

	if err != nil {
		t.Fatalf("RunClarify() error = %v", err)
	}
	if qas != nil {
		t.Errorf("RunClarify() should return nil when no event channel")
	}
}

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
