package review

import (
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/shared/testgit"
)

func setupCleanGitRepo(t *testing.T) string {
	t.Helper()
	workDir := t.TempDir()
	testgit.InitRepo(t, workDir)
	return workDir
}

func setupGitRepoWithWorkingTreeDiff(t *testing.T) (workDir, changedFile string) {
	t.Helper()
	workDir = setupCleanGitRepo(t)
	changedFile = "delta.txt"
	if err := os.WriteFile(filepath.Join(workDir, changedFile), []byte("changed\n"), 0644); err != nil {
		t.Fatalf("write %s: %v", changedFile, err)
	}
	return workDir, changedFile
}
