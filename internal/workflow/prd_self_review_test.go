package workflow

import (
	"context"
	"encoding/json"
	"errors"
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
		Stories:     []*prd.Story{{ID: "1", Title: "Story", Description: "Desc", Slices: testStorySlice("AC"), Priority: 1}},
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

func TestRunPRDSelfReviewMissingVerdictRetriesUntilMaxRounds(t *testing.T) {
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
	if mock.CallCount() != constants.MaxPRDSelfReviewRounds {
		t.Errorf("runner calls = %d, want %d (missing verdict retries until cap)", mock.CallCount(), constants.MaxPRDSelfReviewRounds)
	}

	texts := drainOutputTexts(ch)
	foundExhausted := false
	for _, text := range texts {
		if strings.Contains(text, "did not approve within") {
			foundExhausted = true
		}
	}
	if !foundExhausted {
		t.Errorf("expected exhausted-rounds message in outputs %v", texts)
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

func TestRunPRDSelfReviewRunnerErrorDegradesToCurrentPRD(t *testing.T) {
	cfg := newSelfReviewConfig(t)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		return errors.New("runner exploded")
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	p, err := exec.runPRDSelfReview(context.Background(), "build feature")
	if err != nil {
		t.Fatalf("runPRDSelfReview() error = %v, want graceful degradation", err)
	}
	if p == nil || p.ProjectName != "Test" {
		t.Fatalf("runPRDSelfReview() PRD = %+v, want on-disk PRD", p)
	}
	if mock.CallCount() != 1 {
		t.Errorf("runner calls = %d, want 1 (no retries after runner error)", mock.CallCount())
	}

	texts := drainOutputTexts(ch)
	foundDegraded := false
	for _, text := range texts {
		if strings.Contains(text, "failed, proceeding with current PRD") {
			foundDegraded = true
		}
	}
	if !foundDegraded {
		t.Errorf("expected degradation message in outputs %v", texts)
	}
}

func TestRunPRDSelfReviewCanceledContextReturnsError(t *testing.T) {
	cfg := newSelfReviewConfig(t)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		return ctx.Err()
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	exec := NewExecutorWithRunner(cfg, ch, mock)
	if _, err := exec.runPRDSelfReview(ctx, "build feature"); err == nil {
		t.Fatal("runPRDSelfReview() error = nil, want context cancellation error")
	}
}

func TestRunPRDSelfReviewRemovesStaleVerdict(t *testing.T) {
	cfg := newSelfReviewConfig(t)
	if err := writeVerdictFile(t, cfg.WorkDir, true, "stale approval from a previous run"); err != nil {
		t.Fatalf("failed to seed stale verdict: %v", err)
	}

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
	if p == nil {
		t.Fatal("runPRDSelfReview() returned nil PRD")
	}
	if mock.CallCount() != constants.MaxPRDSelfReviewRounds {
		t.Errorf("runner calls = %d, want %d (stale verdict must not count as round-1 approval)", mock.CallCount(), constants.MaxPRDSelfReviewRounds)
	}
}

func TestRunGenerateWithoutAutoApproveSkipsSelfReview(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if strings.Contains(p, prompt.PRDSelfReviewVerdictFile) {
			t.Errorf("self-review should not run without auto-approve, got prompt:\n%s", p)
			return nil
		}
		data := `{"project_name":"Generated","stories":[{"id":"1","title":"Test","description":"Desc","slices":[{"id":"slice-1","behavior":"AC","red_hint":"add failing test"}],"priority":1}]}`
		return os.WriteFile(filepath.Join(tmpDir, "prd.json"), []byte(data), 0644)
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	p, err := exec.RunGenerate(context.Background(), "build feature")
	if err != nil {
		t.Fatalf("RunGenerate() error = %v", err)
	}
	if p == nil {
		t.Fatal("RunGenerate() returned nil PRD")
	}
	if mock.CallCount() != 1 {
		t.Errorf("runner calls = %d, want 1 (generation only)", mock.CallCount())
	}

	reviewEvents := 0
	for len(ch) > 0 {
		if _, ok := (<-ch).(EventPRDReview); ok {
			reviewEvents++
		}
	}
	if reviewEvents != 1 {
		t.Errorf("EventPRDReview emitted %d times, want 1", reviewEvents)
	}
}

func TestRunGenerateRunsSelfReviewBeforePRDReview(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"
	cfg.AutoApprove = true

	ch := make(chan Event, 100)
	mock := newMockRunner()
	reviewCalls := 0
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if !strings.Contains(p, prompt.PRDSelfReviewVerdictFile) {
			data := `{"project_name":"Generated","stories":[{"id":"1","title":"Test","description":"Desc","slices":[{"id":"slice-1","behavior":"AC","red_hint":"add failing test"}],"priority":1}]}`
			return os.WriteFile(filepath.Join(tmpDir, "prd.json"), []byte(data), 0644)
		}
		reviewCalls++
		if reviewCalls == 1 {
			return writeVerdictFile(t, tmpDir, false, "needs work")
		}
		return writeVerdictFile(t, tmpDir, true, "fixed")
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	p, err := exec.RunGenerate(context.Background(), "build feature")
	if err != nil {
		t.Fatalf("RunGenerate() error = %v", err)
	}
	if p == nil {
		t.Fatal("RunGenerate() returned nil PRD")
	}
	if mock.CallCount() != 3 {
		t.Errorf("runner calls = %d, want 3 (1 generation + 2 review rounds)", mock.CallCount())
	}

	reviewEvents := 0
	for len(ch) > 0 {
		if _, ok := (<-ch).(EventPRDReview); ok {
			reviewEvents++
		}
	}
	if reviewEvents != 1 {
		t.Errorf("EventPRDReview emitted %d times, want 1", reviewEvents)
	}
}

func TestRunGenerateSelfReviewNeverApprovedStillEmitsPRDReview(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"
	cfg.AutoApprove = true

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if !strings.Contains(p, prompt.PRDSelfReviewVerdictFile) {
			data := `{"project_name":"Generated","stories":[{"id":"1","title":"Test","description":"Desc","slices":[{"id":"slice-1","behavior":"AC","red_hint":"add failing test"}],"priority":1}]}`
			return os.WriteFile(filepath.Join(tmpDir, "prd.json"), []byte(data), 0644)
		}
		return writeVerdictFile(t, tmpDir, false, "never satisfied")
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	p, err := exec.RunGenerate(context.Background(), "build feature")
	if err != nil {
		t.Fatalf("RunGenerate() error = %v", err)
	}
	if mock.CallCount() != 1+constants.MaxPRDSelfReviewRounds {
		t.Errorf("runner calls = %d, want %d", mock.CallCount(), 1+constants.MaxPRDSelfReviewRounds)
	}

	reviewEvents := 0
	for len(ch) > 0 {
		if re, ok := (<-ch).(EventPRDReview); ok {
			reviewEvents++
			if re.PRD.ProjectName != p.ProjectName {
				t.Errorf("EventPRDReview PRD = %q, want %q", re.PRD.ProjectName, p.ProjectName)
			}
		}
	}
	if reviewEvents != 1 {
		t.Errorf("EventPRDReview emitted %d times, want 1", reviewEvents)
	}
}

func TestRunGenerateSelfReviewMissingVerdictProceedsToPRDReview(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"
	cfg.AutoApprove = true

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if !strings.Contains(p, prompt.PRDSelfReviewVerdictFile) {
			data := `{"project_name":"Generated","stories":[{"id":"1","title":"Test","description":"Desc","slices":[{"id":"slice-1","behavior":"AC","red_hint":"add failing test"}],"priority":1}]}`
			return os.WriteFile(filepath.Join(tmpDir, "prd.json"), []byte(data), 0644)
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	if _, err := exec.RunGenerate(context.Background(), "build feature"); err != nil {
		t.Fatalf("RunGenerate() error = %v", err)
	}
	if mock.CallCount() != 1+constants.MaxPRDSelfReviewRounds {
		t.Errorf("runner calls = %d, want %d", mock.CallCount(), 1+constants.MaxPRDSelfReviewRounds)
	}

	reviewEvents := 0
	foundExhausted := false
	for len(ch) > 0 {
		switch e := (<-ch).(type) {
		case EventPRDReview:
			reviewEvents++
		case EventOutput:
			if strings.Contains(e.Text, "did not approve within") {
				foundExhausted = true
			}
		}
	}
	if reviewEvents != 1 {
		t.Errorf("EventPRDReview emitted %d times, want 1", reviewEvents)
	}
	if !foundExhausted {
		t.Error("expected exhausted-rounds output before PRD review")
	}
}

func TestRunGenerateSelfReviewErrorSkipsPRDReview(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"
	cfg.AutoApprove = true

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if !strings.Contains(p, prompt.PRDSelfReviewVerdictFile) {
			data := `{"project_name":"Generated","stories":[{"id":"1","title":"Test","description":"Desc","slices":[{"id":"slice-1","behavior":"AC","red_hint":"add failing test"}],"priority":1}]}`
			return os.WriteFile(filepath.Join(tmpDir, "prd.json"), []byte(data), 0644)
		}
		invalid := `{"project_name":"Broken","stories":[{"id":"","title":"No ID"}]}`
		if err := os.WriteFile(filepath.Join(tmpDir, "prd.json"), []byte(invalid), 0644); err != nil {
			return err
		}
		return writeVerdictFile(t, tmpDir, true, "approved a broken prd")
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	_, err := exec.RunGenerate(context.Background(), "build feature")
	if err == nil {
		t.Fatal("RunGenerate() should return error when self-review leaves PRD invalid")
	}

	for len(ch) > 0 {
		if _, ok := (<-ch).(EventPRDReview); ok {
			t.Fatal("EventPRDReview should not be emitted when self-review fails")
		}
	}
}
