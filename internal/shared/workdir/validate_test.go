package workdir_test

import (
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/shared/workdir"
)

func TestValidateGitWorkdirMissingDir(t *testing.T) {
	err := workdir.ValidateGit("/does/not/exist")
	if err == nil {
		t.Fatal("ValidateGit() = nil, want error")
	}
}

func TestValidateGitWorkdirWithoutGit(t *testing.T) {
	dir := t.TempDir()
	if err := workdir.ValidateGit(dir); err == nil {
		t.Fatal("ValidateGit() = nil, want error for non-git dir")
	}
}

func TestValidateGitWorkdirOK(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := workdir.ValidateGit(dir); err != nil {
		t.Fatalf("ValidateGit() = %v, want nil", err)
	}
}
