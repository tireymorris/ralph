package workflow

import (
	"context"
	"strings"
	"testing"

	"github.com/tireymorris/ralph/internal/shared/config"
	"github.com/tireymorris/ralph/internal/shared/prd"
	"github.com/tireymorris/ralph/internal/shared/runner"
	"github.com/tireymorris/ralph/internal/shared/testgit"
)

func TestRunCleanupMultipleRoundsUntilNoProgress(t *testing.T) {
	workDir, changedFile := testgit.RepoWithWorkingTreeDiff(t)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()
	calls := 0
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		if isDiffReviewPrompt(prompt) {
			outputCh <- runner.OutputLine{Text: cleanReviewTranscript}
			return nil
		}
		calls++
		if !strings.Contains(prompt, changedFile) {
			t.Errorf("cleanup prompt should contain changed file %q", changedFile)
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	p := &prd.PRD{Context: "ctx"}

	if err := exec.RunCleanup(context.Background(), p); err != nil {
		t.Fatalf("RunCleanup() error = %v", err)
	}

	if calls != 1 {
		t.Fatalf("runner call count = %d, want 1 when cleanup makes no file changes", calls)
	}

	evts := drainEvents(ch)
	counts := countCleanupEvents(evts)
	if counts.started != 1 || counts.completed != 1 {
		t.Errorf("expected 1 cleanup started and 1 completed, got started=%d completed=%d", counts.started, counts.completed)
	}
}
