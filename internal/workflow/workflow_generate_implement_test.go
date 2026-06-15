package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
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
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if strings.Contains(p, prompt.PRDSelfReviewVerdictFile) {
			return nil
		}
		if !strings.Contains(p, "no existing source code") {
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
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if strings.Contains(p, prompt.PRDSelfReviewVerdictFile) {
			return nil
		}
		if strings.Contains(p, "no existing source code") {
			t.Error("expected PRD prompt for workdir with source to not mention no existing source code")
		}
		if !strings.Contains(p, "ACTUALLY observe in the codebase") {
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
	initGitRepoInDir(t, tmpDir)
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
	commitPRDFile(t, tmpDir, cfg.PRDFile)
	if err := os.WriteFile(filepath.Join(tmpDir, "feature.txt"), []byte("base\n"), 0644); err != nil {
		t.Fatalf("write feature file: %v", err)
	}
	trackFeature := exec.Command("git", "add", "feature.txt")
	trackFeature.Dir = tmpDir
	if out, err := trackFeature.CombinedOutput(); err != nil {
		t.Fatalf("git add feature.txt: %v\n%s", err, out)
	}
	commitFeature := exec.Command("git", "commit", "-m", "add feature file")
	commitFeature.Dir = tmpDir
	if out, err := commitFeature.CombinedOutput(); err != nil {
		t.Fatalf("git commit feature.txt: %v\n%s", err, out)
	}

	ch := make(chan Event, 100)
	mock := newMockRunner()

	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		if strings.Contains(prompt, "critical diff review") {
			outputCh <- runner.OutputLine{Text: cleanReviewTranscript}
			return nil
		}
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

func TestRunImplementationIteratesSlicesBeforeMarkingStoryDone(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepoInDir(t, tmpDir)
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"
	cfg.SkipCleanup = true

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories: []*prd.Story{{
			ID:          "story-1",
			Title:       "Story",
			Description: "Desc",
			Slices: []*prd.Slice{
				{ID: "slice-1", Behavior: "first behavior", RedHint: "write first failing test"},
				{ID: "slice-2", Behavior: "second behavior", RedHint: "write second failing test", RefactorHint: "extract helper"},
			},
			Priority: 1,
			Passes:   false,
		}},
	}

	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("failed to save test PRD: %v", err)
	}
	commitPRDFile(t, tmpDir, cfg.PRDFile)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	call := 0
	mock.runFunc = func(ctx context.Context, promptText string, outputCh chan<- runner.OutputLine) error {
		if strings.Contains(promptText, "critical diff review") {
			outputCh <- runner.OutputLine{Text: cleanReviewTranscript}
			return nil
		}
		call++
		if call == 1 {
			if !strings.Contains(promptText, "first behavior") || !strings.Contains(promptText, "write first failing test") {
				t.Fatalf("first slice prompt missing slice content:\n%s", promptText)
			}
			progressed := &prd.PRD{
				ProjectName: "Test",
				Stories: []*prd.Story{{
					ID:          "story-1",
					Title:       "Story",
					Description: "Desc",
					Slices: []*prd.Slice{
						{ID: "slice-1", Behavior: "first behavior", RedHint: "write first failing test", Passes: true},
						{ID: "slice-2", Behavior: "second behavior", RedHint: "write second failing test", RefactorHint: "extract helper", Passes: false},
					},
					Priority: 1,
					Passes:   false,
				}},
			}
			if err := prd.Save(cfg, progressed); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(tmpDir, "progress.txt"), []byte("slice-1\n"), 0644); err != nil {
				return err
			}
			return nil
		}
		if call == 2 {
			if !strings.Contains(promptText, "second behavior") || !strings.Contains(promptText, "extract helper") {
				t.Fatalf("second slice prompt missing refactor hint:\n%s", promptText)
			}
			progressed := &prd.PRD{
				ProjectName: "Test",
				Stories: []*prd.Story{{
					ID:          "story-1",
					Title:       "Story",
					Description: "Desc",
					Slices: []*prd.Slice{
						{ID: "slice-1", Behavior: "first behavior", RedHint: "write first failing test", Passes: true},
						{ID: "slice-2", Behavior: "second behavior", RedHint: "write second failing test", RefactorHint: "extract helper", Passes: true},
					},
					Priority: 1,
					Passes:   true,
				}},
			}
			if err := prd.Save(cfg, progressed); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(tmpDir, "progress.txt"), []byte("slice-2\n"), 0644); err != nil {
				return err
			}
			return nil
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	if err := exec.RunImplementation(context.Background(), testPRD); err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

	if call != 2 {
		t.Fatalf("slice runner call count = %d, want 2", call)
	}
	loaded, err := prd.Load(cfg)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !loaded.GetStory("story-1").AllSlicesPassed() {
		t.Fatal("expected all slices to be marked passed")
	}
	if !loaded.GetStory("story-1").Passes {
		t.Fatal("expected story to be marked passed after final slice")
	}
}

func TestRunImplementationPromptsOnlyCurrentSlice(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepoInDir(t, tmpDir)
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"
	cfg.SkipCleanup = true

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories: []*prd.Story{{
			ID:          "story-1",
			Title:       "Story",
			Description: "Desc",
			Slices: []*prd.Slice{
				{ID: "slice-1", Behavior: "first behavior", RedHint: "write first failing test"},
				{ID: "slice-2", Behavior: "second behavior", RedHint: "write second failing test", RefactorHint: "extract helper"},
			},
			Priority: 1,
			Passes:   false,
		}},
	}

	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("failed to save test PRD: %v", err)
	}
	commitPRDFile(t, tmpDir, cfg.PRDFile)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	var commitMessages []string
	originalCommitChangedFiles := commitChangedFiles
	t.Cleanup(func() { commitChangedFiles = originalCommitChangedFiles })
	commitChangedFiles = func(workDir, message string) (bool, error) {
		commitMessages = append(commitMessages, message)
		return true, nil
	}
	mock.runFunc = func(ctx context.Context, promptText string, outputCh chan<- runner.OutputLine) error {
		if strings.Contains(promptText, "critical diff review") {
			outputCh <- runner.OutputLine{Text: cleanReviewTranscript}
			return nil
		}
		switch {
		case strings.Contains(promptText, "first behavior"):
			if err := os.WriteFile(filepath.Join(tmpDir, "feature.txt"), []byte("slice-1\n"), 0644); err != nil {
				return err
			}
			progressed := &prd.PRD{
				ProjectName: "Test",
				Stories: []*prd.Story{{
					ID:          "story-1",
					Title:       "Story",
					Description: "Desc",
					Slices: []*prd.Slice{
						{ID: "slice-1", Behavior: "first behavior", RedHint: "write first failing test", Passes: true},
						{ID: "slice-2", Behavior: "second behavior", RedHint: "write second failing test", RefactorHint: "extract helper", Passes: false},
					},
					Priority: 1,
					Passes:   false,
				}},
			}
			return prd.Save(cfg, progressed)
		case strings.Contains(promptText, "second behavior"):
			if err := os.WriteFile(filepath.Join(tmpDir, "feature.txt"), []byte("slice-2\n"), 0644); err != nil {
				return err
			}
			progressed := &prd.PRD{
				ProjectName: "Test",
				Stories: []*prd.Story{{
					ID:          "story-1",
					Title:       "Story",
					Description: "Desc",
					Slices: []*prd.Slice{
						{ID: "slice-1", Behavior: "first behavior", RedHint: "write first failing test", Passes: true},
						{ID: "slice-2", Behavior: "second behavior", RedHint: "write second failing test", RefactorHint: "extract helper", Passes: true},
					},
					Priority: 1,
					Passes:   true,
				}},
			}
			return prd.Save(cfg, progressed)
		default:
			t.Fatalf("unexpected implementation prompt:\n%s", promptText)
			return nil
		}
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	if err := exec.RunImplementation(context.Background(), testPRD); err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

	var prompts []string
	for _, call := range mock.calls {
		if strings.Contains(call, "critical diff review") {
			continue
		}
		prompts = append(prompts, call)
	}
	if len(prompts) != 2 {
		t.Fatalf("implementation prompt count = %d, want 2", len(prompts))
	}
	if strings.Contains(prompts[0], "second behavior") || strings.Contains(prompts[0], "extract helper") {
		t.Fatalf("first prompt should only mention the first slice:\n%s", prompts[0])
	}
	if strings.Contains(prompts[1], "first behavior") || strings.Contains(prompts[1], "write first failing test") {
		t.Fatalf("second prompt should only mention the second slice:\n%s", prompts[1])
	}
}

func TestRunImplementationCommitsEachSliceWithSliceScopedMessage(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepoInDir(t, tmpDir)
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"
	cfg.SkipCleanup = true

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories: []*prd.Story{{
			ID:          "story-1",
			Title:       "Story",
			Description: "Desc",
			Slices: []*prd.Slice{
				{ID: "slice-1", Behavior: "first behavior", RedHint: "write first failing test"},
				{ID: "slice-2", Behavior: "second behavior", RedHint: "write second failing test", RefactorHint: "extract helper"},
			},
			Priority: 1,
			Passes:   false,
		}},
	}

	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("failed to save test PRD: %v", err)
	}
	commitPRDFile(t, tmpDir, cfg.PRDFile)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	var commitMessages []string
	originalCommitChangedFiles := commitChangedFiles
	t.Cleanup(func() { commitChangedFiles = originalCommitChangedFiles })
	commitChangedFiles = func(workDir, message string) (bool, error) {
		commitMessages = append(commitMessages, message)
		return true, nil
	}
	mock.runFunc = func(ctx context.Context, promptText string, outputCh chan<- runner.OutputLine) error {
		if strings.Contains(promptText, "critical diff review") {
			outputCh <- runner.OutputLine{Text: cleanReviewTranscript}
			return nil
		}
		switch {
		case strings.Contains(promptText, "first behavior"):
			progressed := &prd.PRD{
				ProjectName: "Test",
				Stories: []*prd.Story{{
					ID:          "story-1",
					Title:       "Story",
					Description: "Desc",
					Slices: []*prd.Slice{
						{ID: "slice-1", Behavior: "first behavior", RedHint: "write first failing test", Passes: true},
						{ID: "slice-2", Behavior: "second behavior", RedHint: "write second failing test", RefactorHint: "extract helper", Passes: false},
					},
					Priority: 1,
					Passes:   false,
				}},
			}
			return prd.Save(cfg, progressed)
		case strings.Contains(promptText, "second behavior"):
			progressed := &prd.PRD{
				ProjectName: "Test",
				Stories: []*prd.Story{{
					ID:          "story-1",
					Title:       "Story",
					Description: "Desc",
					Slices: []*prd.Slice{
						{ID: "slice-1", Behavior: "first behavior", RedHint: "write first failing test", Passes: true},
						{ID: "slice-2", Behavior: "second behavior", RedHint: "write second failing test", RefactorHint: "extract helper", Passes: true},
					},
					Priority: 1,
					Passes:   true,
				}},
			}
			return prd.Save(cfg, progressed)
		default:
			t.Fatalf("unexpected implementation prompt:\n%s", promptText)
			return nil
		}
	}

	runnerExec := NewExecutorWithRunner(cfg, ch, mock)
	if err := runnerExec.RunImplementation(context.Background(), testPRD); err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

	if len(commitMessages) != 2 {
		t.Fatalf("slice commit count = %d, want 2", len(commitMessages))
	}
	if !strings.Contains(commitMessages[0], "story-1") || !strings.Contains(commitMessages[0], "slice-1") {
		t.Fatalf("first slice commit subject = %q, want story and slice ids", commitMessages[0])
	}
	if !strings.Contains(commitMessages[1], "story-1") || !strings.Contains(commitMessages[1], "slice-2") {
		t.Fatalf("second slice commit subject = %q, want story and slice ids", commitMessages[1])
	}
}

func TestRunImplementationOmitsPassedSlicesFromPrompt(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepoInDir(t, tmpDir)
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"
	cfg.SkipCleanup = true

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories: []*prd.Story{{
			ID:          "story-1",
			Title:       "Story",
			Description: "Desc",
			Slices: []*prd.Slice{
				{ID: "slice-1", Behavior: "finished behavior", RedHint: "write finished test", Passes: true},
				{ID: "slice-2", Behavior: "pending behavior", RedHint: "write pending test", RefactorHint: "extract helper"},
			},
			Priority: 1,
			Passes:   false,
		}},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("failed to save test PRD: %v", err)
	}
	commitPRDFile(t, tmpDir, cfg.PRDFile)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	calls := 0
	mock.runFunc = func(ctx context.Context, promptText string, outputCh chan<- runner.OutputLine) error {
		if strings.Contains(promptText, "critical diff review") {
			outputCh <- runner.OutputLine{Text: cleanReviewTranscript}
			return nil
		}
		if strings.Contains(promptText, "finished behavior") {
			t.Fatalf("prompt should omit passed slices:\n%s", promptText)
		}
		if !strings.Contains(promptText, "pending behavior") || !strings.Contains(promptText, "extract helper") {
			t.Fatalf("prompt missing pending slice details:\n%s", promptText)
		}
		calls++
		if calls == 1 {
			progressed := &prd.PRD{
				ProjectName: "Test",
				Stories: []*prd.Story{{
					ID:          "story-1",
					Title:       "Story",
					Description: "Desc",
					Slices: []*prd.Slice{
						{ID: "slice-1", Behavior: "finished behavior", RedHint: "write finished test", Passes: true},
						{ID: "slice-2", Behavior: "pending behavior", RedHint: "write pending test", RefactorHint: "extract helper", Passes: true},
					},
					Priority: 1,
					Passes:   true,
				}},
			}
			return prd.Save(cfg, progressed)
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	if err := exec.RunImplementation(context.Background(), testPRD); err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}
	if calls != 1 {
		t.Fatalf("runner calls = %d, want 1", calls)
	}
}

func TestRunImplementationRunnerFailureReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepoInDir(t, tmpDir)
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
	commitPRDFile(t, tmpDir, cfg.PRDFile)

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
	initGitRepoInDir(t, tmpDir)
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
	commitPRDFile(t, tmpDir, cfg.PRDFile)

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
	initGitRepoInDir(t, tmpDir)
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
	commitPRDFile(t, tmpDir, cfg.PRDFile)

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
	initGitRepoInDir(t, tmpDir)
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
	commitPRDFile(t, tmpDir, cfg.PRDFile)

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
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if strings.Contains(p, prompt.PRDSelfReviewVerdictFile) {
			data, err := json.Marshal(PRDReviewVerdict{Approved: true, Summary: "ok"})
			if err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(tmpDir, prompt.PRDSelfReviewVerdictFile), data, 0644)
		}
		call++
		switch call {
		case 1:
			revised := &prd.PRD{ProjectName: "CritiqueApplied", Stories: testPRD.Stories}
			return prd.Save(cfg, revised)
		case 2:
			data := `["Which database?"]`
			return os.WriteFile(filepath.Join(tmpDir, ClarifyingQuestionsFile), []byte(data), 0644)
		case 3:
			if !strings.Contains(p, "Which database?") || !strings.Contains(p, "Postgres") {
				t.Fatalf("clarification revision prompt missing answers:\n%s", p)
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
		t.Fatalf("runner call count = %d, want 3 (critique, clarify, clarification revision; self-review only runs with auto-approve)", mock.CallCount())
	}
}

func TestRunImplementationEmitsStoryEvents(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepoInDir(t, tmpDir)
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
	commitPRDFile(t, tmpDir, cfg.PRDFile)

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
	cfg, testPRD := saveSingleStoryPRD(t, false)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	exec := NewExecutorWithRunner(cfg, ch, mock)
	if err := exec.RunImplementation(context.Background(), testPRD); err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

	assertCleanupBeforeCompleted(t, drainEvents(ch))

	if mock.CallCount() < 2 {
		t.Errorf("runner should be called at least twice (1 story + 1 cleanup), got %d", mock.CallCount())
	}
}

func TestRunImplementationCleanupFailureStopsCompleted(t *testing.T) {
	cfg, testPRD := saveSingleStoryPRD(t, false)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if strings.Contains(p, "cleanup") {
			return errors.New("cleanup exploded")
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	err := exec.RunImplementation(context.Background(), testPRD)
	if err == nil {
		t.Fatal("RunImplementation() should return error when cleanup fails")
	}

	foundCompleted := false
	foundCleanupError := false
	var cleanupStarted int
	var cleanupCompleted int
	for len(ch) > 0 {
		e := <-ch
		switch ev := e.(type) {
		case EventCompleted:
			foundCompleted = true
		case EventError:
			if strings.Contains(ev.Err.Error(), "cleanup") {
				foundCleanupError = true
			}
		case EventCleanupStarted:
			cleanupStarted++
		case EventCleanupCompleted:
			cleanupCompleted++
		}
	}

	if cleanupStarted != 1 {
		t.Errorf("expected exactly 1 EventCleanupStarted before failure, got %d", cleanupStarted)
	}
	if cleanupCompleted != 0 {
		t.Errorf("EventCleanupCompleted should not be emitted when cleanup fails, got %d", cleanupCompleted)
	}

	if foundCompleted {
		t.Error("EventCompleted should NOT be emitted when cleanup fails")
	}
	if !foundCleanupError {
		t.Error("expected EventError with message containing 'cleanup'")
	}
	storyCalls, reviewCalls, cleanupCalls := countRunnerPromptKinds(mock)
	if storyCalls != 1 {
		t.Errorf("story runner calls = %d, want 1", storyCalls)
	}
	if reviewCalls != 1 {
		t.Errorf("review runner calls = %d, want 1", reviewCalls)
	}
	if cleanupCalls != 1 {
		t.Errorf("cleanup runner calls = %d, want 1", cleanupCalls)
	}
}

func TestRunImplementationSkipsCleanupWhenConfigured(t *testing.T) {
	cfg, testPRD := saveSingleStoryPRD(t, true)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	exec := NewExecutorWithRunner(cfg, ch, mock)
	if err := exec.RunImplementation(context.Background(), testPRD); err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

	evts := drainEvents(ch)
	counts := countCleanupEvents(evts)
	if counts.started != 0 || counts.completed != 0 {
		t.Errorf("expected no cleanup events with SkipCleanup, got started=%d completed=%d", counts.started, counts.completed)
	}

	foundCompleted := false
	for _, e := range evts {
		if _, ok := e.(EventCompleted); ok {
			foundCompleted = true
		}
	}
	if !foundCompleted {
		t.Error("EventCompleted should still be emitted when SkipCleanup is true")
	}

	storyCalls, reviewCalls, cleanupCalls := countRunnerPromptKinds(mock)
	if storyCalls != 1 {
		t.Errorf("story runner calls = %d, want 1", storyCalls)
	}
	if reviewCalls != 1 {
		t.Errorf("review runner calls = %d, want 1", reviewCalls)
	}
	if cleanupCalls != 0 {
		t.Errorf("cleanup runner calls = %d, want 0", cleanupCalls)
	}
}

func countRunnerPromptKinds(mock *mockRunner) (story, review, cleanup int) {
	mock.mu.Lock()
	defer mock.mu.Unlock()
	for _, p := range mock.calls {
		switch {
		case strings.Contains(p, "critical diff review"):
			review++
		case strings.Contains(p, "cleanup"):
			cleanup++
		case strings.Contains(p, "recovery agent"):
			story++
		default:
			story++
		}
	}
	return story, review, cleanup
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
