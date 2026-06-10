package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ralph/internal/shared/config"
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
}

func TestMockRunnerImplDelayDoesNotSlowSelfReview(t *testing.T) {
	t.Setenv("RALPH_MOCK_IMPL_DELAY_MS", "500")

	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	cfg.Runner = "mock"

	r := NewMock(cfg)
	ch := make(chan OutputLine, 4)
	start := time.Now()
	if err := r.Run(context.Background(), "Review prd and write .ralph_prd_review.json", ch); err != nil {
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
