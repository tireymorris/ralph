package workflow

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
)

func newSelfReviewConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	cfg.PRDFile = "prd.json"

	seeded := &prd.PRD{
		ProjectName: "Test",
		Stories:     []*prd.Story{{ID: "1", Title: "Story", Description: "Desc", AcceptanceCriteria: []string{"AC"}, Priority: 1}},
	}
	if err := prd.Save(cfg, seeded); err != nil {
		t.Fatalf("failed to seed PRD: %v", err)
	}
	return cfg
}

func writeVerdictFile(t *testing.T, workDir string, approved bool, summary string) error {
	t.Helper()
	data, err := json.Marshal(PRDReviewVerdict{Approved: approved, Summary: summary})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(workDir, prompt.PRDSelfReviewVerdictFile), data, 0644)
}

func drainOutputTexts(ch chan Event) []string {
	var texts []string
	for len(ch) > 0 {
		if eo, ok := (<-ch).(EventOutput); ok {
			texts = append(texts, eo.Text)
		}
	}
	return texts
}

func TestRunPRDSelfReviewApprovedFirstRound(t *testing.T) {
	cfg := newSelfReviewConfig(t)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if !strings.Contains(p, prompt.PRDSelfReviewVerdictFile) {
			t.Errorf("self-review prompt should mention verdict file, got:\n%s", p)
		}
		return writeVerdictFile(t, cfg.WorkDir, true, "rubric satisfied")
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	p, err := exec.runPRDSelfReview(context.Background(), "build feature")
	if err != nil {
		t.Fatalf("runPRDSelfReview() error = %v", err)
	}
	if p == nil || p.ProjectName != "Test" {
		t.Fatalf("runPRDSelfReview() PRD = %+v, want reloaded PRD", p)
	}
	if mock.CallCount() != 1 {
		t.Errorf("runner calls = %d, want 1", mock.CallCount())
	}

	texts := drainOutputTexts(ch)
	foundRound := false
	foundSummary := false
	for _, text := range texts {
		if strings.Contains(text, "PRD self-review round 1 of") {
			foundRound = true
		}
		if strings.Contains(text, "rubric satisfied") {
			foundSummary = true
		}
	}
	if !foundRound {
		t.Errorf("expected round announcement in outputs %v", texts)
	}
	if !foundSummary {
		t.Errorf("expected verdict summary in outputs %v", texts)
	}
	if _, statErr := os.Stat(filepath.Join(cfg.WorkDir, prompt.PRDSelfReviewVerdictFile)); !os.IsNotExist(statErr) {
		t.Error("expected verdict file to be removed after read")
	}
}

func TestRunPRDSelfReviewLoopsUntilApproved(t *testing.T) {
	cfg := newSelfReviewConfig(t)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	call := 0
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		call++
		if call == 1 {
			return writeVerdictFile(t, cfg.WorkDir, false, "needs work")
		}
		return writeVerdictFile(t, cfg.WorkDir, true, "fixed")
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	p, err := exec.runPRDSelfReview(context.Background(), "build feature")
	if err != nil {
		t.Fatalf("runPRDSelfReview() error = %v", err)
	}
	if p == nil {
		t.Fatal("runPRDSelfReview() returned nil PRD")
	}
	if mock.CallCount() != 2 {
		t.Errorf("runner calls = %d, want 2", mock.CallCount())
	}

	texts := drainOutputTexts(ch)
	foundRound2 := false
	for _, text := range texts {
		if strings.Contains(text, "PRD self-review round 2 of") {
			foundRound2 = true
		}
	}
	if !foundRound2 {
		t.Errorf("expected round 2 announcement in outputs %v", texts)
	}
}

func TestRunPRDSelfReviewStopsAtMaxRounds(t *testing.T) {
	cfg := newSelfReviewConfig(t)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		return writeVerdictFile(t, cfg.WorkDir, false, "still not good enough")
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	p, err := exec.runPRDSelfReview(context.Background(), "build feature")
	if err != nil {
		t.Fatalf("runPRDSelfReview() error = %v", err)
	}
	if p == nil || p.ProjectName != "Test" {
		t.Fatalf("runPRDSelfReview() PRD = %+v, want last valid PRD", p)
	}
	if mock.CallCount() != constants.MaxPRDSelfReviewRounds {
		t.Errorf("runner calls = %d, want %d", mock.CallCount(), constants.MaxPRDSelfReviewRounds)
	}
}

func TestRunPRDSelfReviewMissingVerdictCountsAsApproved(t *testing.T) {
	cfg := newSelfReviewConfig(t)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	p, err := exec.runPRDSelfReview(context.Background(), "build feature")
	if err != nil {
		t.Fatalf("runPRDSelfReview() error = %v", err)
	}
	if p == nil || p.ProjectName != "Test" {
		t.Fatalf("runPRDSelfReview() PRD = %+v, want reloaded PRD", p)
	}
	if mock.CallCount() != 1 {
		t.Errorf("runner calls = %d, want 1 (missing verdict counts as approved)", mock.CallCount())
	}
}

func TestRunPRDSelfReviewInvalidPRDReturnsValidationError(t *testing.T) {
	cfg := newSelfReviewConfig(t)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		invalid := `{"project_name":"Broken","stories":[{"id":"","title":"No ID"}]}`
		if err := os.WriteFile(filepath.Join(cfg.WorkDir, "prd.json"), []byte(invalid), 0644); err != nil {
			return err
		}
		return writeVerdictFile(t, cfg.WorkDir, true, "approved a broken prd")
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	p, err := exec.runPRDSelfReview(context.Background(), "build feature")
	if err == nil {
		t.Fatal("runPRDSelfReview() should return error when revised PRD fails validation")
	}
	if p != nil {
		t.Errorf("runPRDSelfReview() PRD = %+v, want nil on validation error", p)
	}
}
