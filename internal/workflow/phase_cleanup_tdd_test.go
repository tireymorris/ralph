package workflow

import (
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/shared/gitdiff"
)

func TestBranchChangedFilesIncludesWorktreeChanges(t *testing.T) {
	workDir := setupCleanupBranchWithUpstreamDiff(t)
	created := "worktree-added.txt"
	if err := os.WriteFile(filepath.Join(workDir, created), []byte("created during cleanup\n"), 0644); err != nil {
		t.Fatalf("write worktree file: %v", err)
	}

	got, err := gitdiff.ChangedFiles(workDir)
	if err != nil {
		t.Fatalf("ChangedFiles() err = %v", err)
	}
	for _, name := range got {
		if name == created {
			return
		}
	}
	t.Fatalf("ChangedFiles() = %v, want to include %q", got, created)
}
