package workdir_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"ralph/internal/shared/workdir"
)

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "t@example.com"},
		{"config", "user.name", "t"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func TestCurrentBranchNameFeatureBranch(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", "README.md"},
		{"commit", "-m", "init"},
		{"checkout", "-b", "feature/test"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	got, err := workdir.CurrentBranchName(dir)
	if err != nil {
		t.Fatalf("CurrentBranchName() error = %v", err)
	}
	if got != "feature/test" {
		t.Fatalf("CurrentBranchName() = %q, want %q", got, "feature/test")
	}
}

func TestIsDefaultBranch(t *testing.T) {
	defaults := []string{"main", "master", "develop", "trunk"}
	tests := []struct {
		name       string
		branchName string
		want       bool
	}{
		{name: "main", branchName: "main", want: true},
		{name: "master", branchName: "master", want: true},
		{name: "develop", branchName: "develop", want: true},
		{name: "trunk", branchName: "trunk", want: true},
		{name: "feature branch", branchName: "feature/test", want: false},
		{name: "release branch", branchName: "release/1.0", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := workdir.IsDefaultBranch(tt.branchName, defaults); got != tt.want {
				t.Fatalf("IsDefaultBranch(%q) = %v, want %v", tt.branchName, got, tt.want)
			}
		})
	}
}
