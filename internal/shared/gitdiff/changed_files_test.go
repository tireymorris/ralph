package gitdiff

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestHashFilesOrderInvariant(t *testing.T) {
	a := HashFiles([]string{"b.go", "a.go"})
	b := HashFiles([]string{"a.go", "b.go"})
	if a != b {
		t.Fatalf("HashFiles() = %q vs %q", a, b)
	}
}

func TestChangedFilesNonGitWorkdir(t *testing.T) {
	_, err := ChangedFiles(t.TempDir())
	if err == nil {
		t.Fatal("ChangedFiles() err = nil, want GitError")
	}
	var ge *GitError
	if !errors.As(err, &ge) {
		t.Fatalf("err = %T %v, want *GitError", err, err)
	}
}

func TestChangedFilesIncludesWorktreeFile(t *testing.T) {
	workDir := t.TempDir()
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
	if err := os.WriteFile(filepath.Join(workDir, "base.txt"), []byte("base\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit("add", "base.txt")
	runGit("commit", "-m", "base")

	created := "delta.txt"
	if err := os.WriteFile(filepath.Join(workDir, created), []byte("new\n"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := ChangedFiles(workDir)
	if err != nil {
		t.Fatalf("ChangedFiles() err = %v", err)
	}
	for _, name := range got {
		if name == created {
			return
		}
	}
	t.Fatalf("ChangedFiles() = %v, want %q", got, created)
}

func TestExcludeReviewArtifactsOmitsRalphRuntimeFiles(t *testing.T) {
	files := []string{
		"hello.txt",
		"prd.json",
		"prd.json.lock",
		".ralph/runs/x/meta.json",
		"tmp/scratch.txt",
	}
	got := ExcludeReviewArtifacts(files)
	want := []string{"hello.txt", "prd.json", "tmp/scratch.txt"}
	if len(got) != len(want) {
		t.Fatalf("ExcludeReviewArtifacts() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ExcludeReviewArtifacts() = %v, want %v", got, want)
		}
	}
}
