package workflow

import (
	"os"
	"path/filepath"
	"testing"
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
