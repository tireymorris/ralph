package testgit

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitRepoCreatesMainBranchWithInitialCommit(t *testing.T) {
	dir := t.TempDir()
	InitRepo(t, dir)

	out, err := exec.Command("git", "-C", dir, "branch", "--show-current").CombinedOutput()
	if err != nil {
		t.Fatalf("git branch --show-current: %v\n%s", err, out)
	}
	if branch := strings.TrimSpace(string(out)); branch != "main" {
		t.Fatalf("branch = %q, want main", branch)
	}

	out, err = exec.Command("git", "-C", dir, "log", "-1", "--oneline").CombinedOutput()
	if err != nil {
		t.Fatalf("git log: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "initial") {
		t.Fatalf("log = %q, want commit message containing initial", out)
	}
}

func TestCommitFileCreatesCommit(t *testing.T) {
	dir := t.TempDir()
	InitRepo(t, dir)

	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write notes.txt: %v", err)
	}

	CommitFile(t, dir, "notes.txt", "add notes")

	out, err := exec.Command("git", "-C", dir, "log", "-1", "--oneline").CombinedOutput()
	if err != nil {
		t.Fatalf("git log: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "add notes") {
		t.Fatalf("log = %q, want commit message containing add notes", out)
	}
}

func TestWriteFileCreatesFile(t *testing.T) {
	dir := t.TempDir()
	WriteFile(t, dir, "notes.txt", "hello\n")

	got, err := os.ReadFile(filepath.Join(dir, "notes.txt"))
	if err != nil {
		t.Fatalf("read notes.txt: %v", err)
	}
	if string(got) != "hello\n" {
		t.Fatalf("notes.txt = %q, want hello\\n", got)
	}
}

func TestRepoWithWorkingTreeDiffReturnsDirtyRepo(t *testing.T) {
	workDir, changedFile := RepoWithWorkingTreeDiff(t)

	out, err := exec.Command("git", "-C", workDir, "branch", "--show-current").CombinedOutput()
	if err != nil {
		t.Fatalf("git branch --show-current: %v\n%s", err, out)
	}
	if branch := strings.TrimSpace(string(out)); branch != "main" {
		t.Fatalf("branch = %q, want main", branch)
	}

	got, err := os.ReadFile(filepath.Join(workDir, changedFile))
	if err != nil {
		t.Fatalf("read %s: %v", changedFile, err)
	}
	if string(got) != "changed\n" {
		t.Fatalf("%s = %q, want changed\\n", changedFile, got)
	}
}
