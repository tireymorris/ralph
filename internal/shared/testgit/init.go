package testgit

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

func InitRepo(t *testing.T, dir string) {
	t.Helper()
	runGit(t, dir, "init")
	runGit(t, dir, "checkout", "-b", "main")
	runGit(t, dir, "config", "user.name", "Test User")
	runGit(t, dir, "config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("ok\n"), 0644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "initial")
}

func CommitFile(t *testing.T, dir, relPath, message string) {
	t.Helper()
	runGit(t, dir, "add", relPath)
	runGit(t, dir, "commit", "-m", message)
}

func WriteFile(t *testing.T, dir, name, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func RepoWithWorkingTreeDiff(t *testing.T) (workDir, changedFile string) {
	t.Helper()
	workDir = t.TempDir()
	InitRepo(t, workDir)
	changedFile = "delta.txt"
	WriteFile(t, workDir, changedFile, "changed\n")
	return workDir, changedFile
}
