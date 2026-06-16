package workdir_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"ralph/internal/shared/workdir"
)

func TestDetectDefaultBranchesIncludesGitOriginHEAD(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", "README.md"},
		{"commit", "-m", "init"},
		{"branch", "-M", "trunk"},
		{"symbolic-ref", "HEAD", "refs/remotes/origin/trunk"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	got := workdir.DetectDefaultBranches(dir)
	for _, name := range got {
		if name == "trunk" {
			return
		}
	}
	t.Fatalf("DetectDefaultBranches() = %v, want to include trunk from origin HEAD", got)
}
