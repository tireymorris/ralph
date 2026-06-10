package gitdiff

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func initGitRepoForCommit(t *testing.T, workDir string) {
	t.Helper()
	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = workDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	runGit("init")
	runGit("checkout", "-b", "main")
	runGit("config", "user.name", "Test User")
	runGit("config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(workDir, "base.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit("add", "base.txt")
	runGit("commit", "-m", "base")
}

func TestCommitChangedFilesCommitsUncommittedWork(t *testing.T) {
	workDir := t.TempDir()
	initGitRepoForCommit(t, workDir)

	if err := os.WriteFile(filepath.Join(workDir, "feature.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	committed, err := CommitChangedFiles(workDir, "ralph: feature")
	if err != nil {
		t.Fatalf("CommitChangedFiles() err = %v", err)
	}
	if !committed {
		t.Fatal("CommitChangedFiles() committed = false, want true")
	}

	status := exec.Command("git", "status", "--porcelain")
	status.Dir = workDir
	out, err := status.Output()
	if err != nil {
		t.Fatal(err)
	}
	if got := string(out); got != "" {
		t.Fatalf("git status after commit = %q, want clean", got)
	}
}

func TestCommitChangedFilesSkipsRalphState(t *testing.T) {
	workDir := t.TempDir()
	initGitRepoForCommit(t, workDir)

	if err := os.MkdirAll(filepath.Join(workDir, ".ralph", "runs", "x"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workDir, ".ralph", "runs", "x", "meta.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	committed, err := CommitChangedFiles(workDir, "ralph: empty")
	if err != nil {
		t.Fatalf("CommitChangedFiles() err = %v", err)
	}
	if committed {
		t.Fatal("CommitChangedFiles() committed = true, want false for ralph-only changes")
	}
}
