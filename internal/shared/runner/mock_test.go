package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
)

func TestMockRunnerWritesPRD(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.Runner = "mock"
	cfg.PRDFile = "prd.json"

	r := NewMock(cfg)
	ch := make(chan OutputLine, 4)
	if err := r.Run(context.Background(), "Generate a PRD.\n3. Write the PRD file, then STOP.", ch); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, "prd.json")); err != nil {
		t.Fatalf("prd.json not written: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(workDir, "prd.json"))
	if err != nil {
		t.Fatalf("failed to read prd.json: %v", err)
	}
	got := string(data)
	for _, want := range []string{`"slices"`, `"behavior"`, `"red_hint"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("prd.json missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "acceptance_criteria") {
		t.Fatalf("prd.json should not contain acceptance_criteria:\n%s", got)
	}
}

func TestMockRunnerImplDelayDoesNotSlowSelfReview(t *testing.T) {
	t.Setenv("RALPH_MOCK_IMPL_DELAY_MS", "500")

	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	cfg.Runner = "mock"

	r := NewMock(cfg)
	ch := make(chan OutputLine, 4)
	start := time.Now()
	if err := r.Run(context.Background(), "Review prd and write .ralph/prd_review.json", ch); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("self-review took %s with implementation delay configured", elapsed)
	}
}

func TestMockRunnerOutputsCleanReviewFindings(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	cfg.Runner = "mock"

	r := NewMock(cfg)
	ch := make(chan OutputLine, 4)
	if err := r.Run(context.Background(), "Review diff.\n===ralph-findings===\n[]\n===/ralph-findings===", ch); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var transcript strings.Builder
	for i := 0; i < 2; i++ {
		transcript.WriteString((<-ch).Text)
		transcript.WriteByte('\n')
	}
	if got := transcript.String(); !strings.Contains(got, "===ralph-findings===\n[]\n===/ralph-findings===") {
		t.Fatalf("transcript missing clean findings block:\n%s", got)
	}
}

func TestMockRunnerAdvancesOneSliceAtATime(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	cfg.Runner = "mock"

	testPRD := &prd.PRD{
		ProjectName: "Mock",
		Stories: []*prd.Story{{
			ID:          "story-1",
			Title:       "Story",
			Description: "Desc",
			Slices: []*prd.Slice{
				{ID: "slice-1", Behavior: "first behavior", RedHint: "write first failing test"},
				{ID: "slice-2", Behavior: "second behavior", RedHint: "write second failing test"},
			},
			Priority: 1,
			Passes:   false,
		}},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	r := NewMock(cfg)
	ch := make(chan OutputLine, 2)

	if err := r.Run(context.Background(), "Implement story: story-1", ch); err != nil {
		t.Fatalf("first Run() error = %v", err)
	}
	got, err := prd.Load(cfg)
	if err != nil {
		t.Fatalf("Load() after first run error = %v", err)
	}
	story := got.GetStory("story-1")
	if story == nil {
		t.Fatal("story missing after first run")
	}
	if !story.Slices[0].Passes || story.Slices[1].Passes || story.Passes {
		t.Fatalf("after first run, slice/story passes = %v/%v/%v, want true/false/false", story.Slices[0].Passes, story.Slices[1].Passes, story.Passes)
	}

	if err := r.Run(context.Background(), "Implement story: story-1", ch); err != nil {
		t.Fatalf("second Run() error = %v", err)
	}
	got, err = prd.Load(cfg)
	if err != nil {
		t.Fatalf("Load() after second run error = %v", err)
	}
	story = got.GetStory("story-1")
	if story == nil {
		t.Fatal("story missing after second run")
	}
	if !story.Slices[0].Passes || !story.Slices[1].Passes || !story.Passes {
		t.Fatalf("after second run, slice/story passes = %v/%v/%v, want true/true/true", story.Slices[0].Passes, story.Slices[1].Passes, story.Passes)
	}
}
