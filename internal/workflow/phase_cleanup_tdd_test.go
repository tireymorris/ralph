package workflow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
)

func TestBranchChangedFilesIncludesWorktreeChanges(t *testing.T) {
	workDir := setupCleanupBranchWithUpstreamDiff(t)
	created := "worktree-added.txt"
	if err := os.WriteFile(filepath.Join(workDir, created), []byte("created during cleanup\n"), 0644); err != nil {
		t.Fatalf("write worktree file: %v", err)
	}

	got := branchChangedFiles(workDir)
	for _, name := range got {
		if name == created {
			return
		}
	}
	t.Fatalf("branchChangedFiles() = %v, want to include %q", got, created)
}

func TestRunCleanupRefreshesChangedFilesBetweenPasses(t *testing.T) {
	workDir := setupCleanupBranchWithUpstreamDiff(t)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()
	call := 0
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		call++
		switch call {
		case 1:
			if !strings.Contains(prompt, "existing-change.txt") {
				t.Fatalf("pass 1 prompt missing upstream diff file:\n%s", prompt)
			}
			if err := os.WriteFile(filepath.Join(workDir, "cleanup-created-by-pass-1.txt"), []byte("new file\n"), 0644); err != nil {
				return err
			}
		case 2:
			if !strings.Contains(prompt, "cleanup-created-by-pass-1.txt") {
				t.Fatalf("pass 2 prompt missing worktree file created after pass 1:\n%s", prompt)
			}
		}
		outputCh <- runner.OutputLine{Text: "refactoring..."}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	if err := exec.RunCleanup(context.Background(), &prd.PRD{Context: "project context"}); err != nil {
		t.Fatalf("RunCleanup() error = %v", err)
	}

	if mock.CallCount() != 3 {
		t.Fatalf("runner call count = %d, want 3", mock.CallCount())
	}
	assertCleanupPassSequence(t, drainEvents(ch), 3)
}

func TestRunCleanupStopsBetweenPassesWhenContextCancelled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()
	ctx, cancel := context.WithCancel(context.Background())
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		cancel()
		outputCh <- runner.OutputLine{Text: "done"}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	err := exec.RunCleanup(ctx, &prd.PRD{Context: "ctx"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunCleanup() error = %v, want context.Canceled", err)
	}

	if mock.CallCount() != 1 {
		t.Fatalf("runner call count = %d, want 1", mock.CallCount())
	}
	assertCleanupPassSequence(t, drainEvents(ch), 1)
}
