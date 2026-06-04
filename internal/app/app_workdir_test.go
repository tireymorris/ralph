package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunRejectsNonGitWorkdirBeforeTUI(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	if code := Run([]string{"--dry-run", "build feature"}); code != 1 {
		t.Fatalf("Run(--dry-run) = %d, want 1 for non-git workdir", code)
	}
}

func TestRunResumeRejectsNonGitWorkdir(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	prd := `{"version":1,"project_name":"p","branch_name":"b","stories":[{"id":"s1","title":"t","description":"d","acceptance_criteria":["a"],"priority":1}]}`
	if err := os.WriteFile(filepath.Join(tmpDir, "prd.json"), []byte(prd), 0644); err != nil {
		t.Fatal(err)
	}

	if code := Run([]string{"--resume", "--dry-run"}); code != 1 {
		t.Fatalf("Run(--resume --dry-run) = %d, want 1 for non-git workdir", code)
	}
}
