package workflow

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func setupGitRepoWithWorkingTreeDiff(t *testing.T) (workDir, changedFile string) {
	t.Helper()
	workDir = t.TempDir()
	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = workDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	runGit("init")
	runGit("checkout", "-b", "main")
	runGit("config", "user.name", "Test User")
	runGit("config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(workDir, "README.md"), []byte("ok\n"), 0644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	runGit("add", "README.md")
	runGit("commit", "-m", "initial")
	changedFile = "delta.txt"
	if err := os.WriteFile(filepath.Join(workDir, changedFile), []byte("changed\n"), 0644); err != nil {
		t.Fatalf("write %s: %v", changedFile, err)
	}
	return workDir, changedFile
}
