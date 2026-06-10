package workflow

import (
	"context"
	"strings"
	"testing"

	"ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
)

func newCritiqueConfig(t *testing.T) *config.Config {
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

func TestRunCritiqueRevisionRunsSelfReviewBeforePRDReview(t *testing.T) {
	cfg := newCritiqueConfig(t)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	reviewCalls := 0
	mock.runFunc = func(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if !strings.Contains(p, prompt.PRDSelfReviewVerdictFile) {
			return nil
		}
		reviewCalls++
		if reviewCalls == 1 {
			return writeVerdictFile(t, cfg.WorkDir, false, "needs work")
		}
		revised := &prd.PRD{
			ProjectName: "PostReview",
			Stories:     []*prd.Story{{ID: "1", Title: "Story", Description: "Desc", AcceptanceCriteria: []string{"AC"}, Priority: 1}},
		}
		if err := prd.Save(cfg, revised); err != nil {
			return err
		}
		return writeVerdictFile(t, cfg.WorkDir, true, "fixed")
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	if err := exec.RunCritiqueRevision(context.Background(), "build feature", "needs more tests"); err != nil {
		t.Fatalf("RunCritiqueRevision() error = %v", err)
	}

	if reviewCalls != 2 {
		t.Errorf("self-review runner calls = %d, want 2", reviewCalls)
	}

	reviewEvents := 0
	for len(ch) > 0 {
		if re, ok := (<-ch).(EventPRDReview); ok {
			reviewEvents++
			if re.PRD.ProjectName != "PostReview" {
				t.Errorf("EventPRDReview project = %q, want PostReview (post-self-review PRD)", re.PRD.ProjectName)
			}
		}
	}
	if reviewEvents != 1 {
		t.Errorf("EventPRDReview emitted %d times, want 1", reviewEvents)
	}
}
