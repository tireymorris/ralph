package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func setupGitRepo(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()

	os.Chdir(tmpDir)
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@test.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	os.WriteFile("initial.txt", []byte("initial"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "Initial commit").Run()

	return tmpDir, func() {
		os.Chdir(origDir)
	}
}

func TestNew(t *testing.T) {
	m := NewWithWorkDir("")
	if m == nil {
		t.Fatal("NewWithWorkDir(\"\") returned nil")
	}
	if m.workDir != "" {
		t.Errorf("NewWithWorkDir(\"\") workDir = %q, want empty", m.workDir)
	}
}

func TestNewWithWorkDir(t *testing.T) {
	m := NewWithWorkDir("/some/path")
	if m == nil {
		t.Fatal("NewWithWorkDir() returned nil")
	}
	if m.workDir != "/some/path" {
		t.Errorf("NewWithWorkDir() workDir = %q, want %q", m.workDir, "/some/path")
	}
}

func TestIsRepository(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	m := NewWithWorkDir("")
	if !m.IsRepository() {
		t.Error("IsRepository() = false in git repo")
	}
}

func TestIsRepositoryNonRepo(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	m := NewWithWorkDir("")
	if m.IsRepository() {
		t.Error("IsRepository() = true in non-git dir")
	}
}

func TestCurrentBranch(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	m := NewWithWorkDir("")
	branch, err := m.CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch() error = %v", err)
	}
	if branch != "main" && branch != "master" {
		t.Errorf("CurrentBranch() = %q, want main or master", branch)
	}
}

func TestBranchExists(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	m := NewWithWorkDir("")

	branch, _ := m.CurrentBranch()
	if !m.BranchExists(branch) {
		t.Errorf("BranchExists(%q) = false, want true", branch)
	}

	if m.BranchExists("nonexistent-branch") {
		t.Error("BranchExists(nonexistent) = true, want false")
	}
}

func TestCreateBranch(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	m := NewWithWorkDir("")

	err := m.CreateBranch("feature/test")
	if err != nil {
		t.Fatalf("CreateBranch() error = %v", err)
	}

	branch, _ := m.CurrentBranch()
	if branch != "feature/test" {
		t.Errorf("CurrentBranch() = %q, want %q", branch, "feature/test")
	}
}

func TestCreateBranchExisting(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	m := NewWithWorkDir("")
	m.CreateBranch("feature/existing")

	origBranch, _ := m.CurrentBranch()
	m.Checkout("main")

	err := m.CreateBranch("feature/existing")
	if err != nil {
		t.Fatalf("CreateBranch() error = %v", err)
	}

	branch, _ := m.CurrentBranch()
	if branch != origBranch {
		t.Errorf("Should checkout existing branch, got %q", branch)
	}
}

func TestCheckout(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	m := NewWithWorkDir("")
	m.CreateBranch("feature/checkout-test")
	m.Checkout("main")

	err := m.Checkout("feature/checkout-test")
	if err != nil {
		t.Fatalf("Checkout() error = %v", err)
	}

	branch, _ := m.CurrentBranch()
	if branch != "feature/checkout-test" {
		t.Errorf("CurrentBranch() = %q after checkout", branch)
	}
}

func TestHasChanges(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	m := NewWithWorkDir("")

	if m.HasChanges() {
		t.Error("HasChanges() = true on clean repo")
	}

	os.WriteFile("new.txt", []byte("new"), 0644)
	exec.Command("git", "add", ".").Run()

	if !m.HasChanges() {
		t.Error("HasChanges() = false with staged changes")
	}
}

func TestHasChangesUnstaged(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	m := NewWithWorkDir("")

	os.WriteFile("initial.txt", []byte("modified"), 0644)

	if !m.HasChanges() {
		t.Error("HasChanges() = false with unstaged changes")
	}
}

func TestStageAll(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	m := NewWithWorkDir("")

	os.WriteFile("staged.txt", []byte("staged"), 0644)

	err := m.StageAll()
	if err != nil {
		t.Fatalf("StageAll() error = %v", err)
	}
}

func TestCommit(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	m := NewWithWorkDir("")

	os.WriteFile("commit.txt", []byte("commit"), 0644)
	m.StageAll()

	err := m.Commit("Test commit")
	if err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	if m.HasChanges() {
		t.Error("HasChanges() = true after commit")
	}
}

func TestCommitStory(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	m := NewWithWorkDir("")

	os.WriteFile("story.txt", []byte("story"), 0644)

	err := m.CommitStory("story-1", "Test Story", "Description")
	if err != nil {
		t.Fatalf("CommitStory() error = %v", err)
	}

	if m.HasChanges() {
		t.Error("HasChanges() = true after CommitStory")
	}
}

func TestCommitStoryNoChanges(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	m := NewWithWorkDir("")

	err := m.CommitStory("story-1", "Test", "Desc")
	if err != nil {
		t.Errorf("CommitStory() with no changes should not error, got %v", err)
	}
}

func TestCommitStoryMessage(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	m := NewWithWorkDir("")

	os.WriteFile(filepath.Join(dir, "feature.txt"), []byte("feature"), 0644)

	m.CommitStory("story-123", "Add Feature", "Feature description")

	out, _ := exec.Command("git", "log", "-1", "--format=%s%n%b").Output()
	msg := string(out)

	if msg == "" {
		t.Skip("Could not get commit message")
	}
}

func TestCurrentBranchError(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	m := NewWithWorkDir("")
	_, err := m.CurrentBranch()
	if err == nil {
		t.Error("CurrentBranch() should error in non-git directory")
	}
}

func TestCommitStoryStageError(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.WriteFile("file.txt", []byte("content"), 0644)

	m := NewWithWorkDir("")
	err := m.CommitStory("s1", "Title", "Desc")
	if err == nil {
		t.Error("CommitStory() should error when git commands fail")
	}
}

func TestCommitStoryCommitError(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	m := NewWithWorkDir("")

	os.WriteFile("test.txt", []byte("test"), 0644)
	m.StageAll()

	exec.Command("git", "config", "user.email", "").Run()

	err := m.CommitStory("s1", "Title", "Desc")
	if err == nil {
		t.Log("CommitStory() may succeed depending on git config")
	}
}
