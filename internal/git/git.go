package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Manager handles git operations
type Manager struct{}

func New() *Manager {
	return &Manager{}
}

// IsRepository checks if the current directory is a git repository
func (m *Manager) IsRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// CurrentBranch returns the current branch name
func (m *Manager) CurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// BranchExists checks if a branch exists
func (m *Manager) BranchExists(name string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", fmt.Sprintf("refs/heads/%s", name))
	return cmd.Run() == nil
}

// CreateBranch creates and switches to a new branch
func (m *Manager) CreateBranch(name string) error {
	if m.BranchExists(name) {
		return m.Checkout(name)
	}
	cmd := exec.Command("git", "checkout", "-b", name)
	return cmd.Run()
}

// Checkout switches to an existing branch
func (m *Manager) Checkout(name string) error {
	cmd := exec.Command("git", "checkout", name)
	return cmd.Run()
}

// HasChanges returns true if there are uncommitted changes
func (m *Manager) HasChanges() bool {
	// Check unstaged
	cmd := exec.Command("git", "diff", "--quiet", "--exit-code")
	if cmd.Run() != nil {
		return true
	}
	// Check staged
	cmd = exec.Command("git", "diff", "--staged", "--quiet", "--exit-code")
	return cmd.Run() != nil
}

// StageAll stages all changes
func (m *Manager) StageAll() error {
	cmd := exec.Command("git", "add", ".")
	return cmd.Run()
}

// Commit creates a commit with the given message
func (m *Manager) Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	return cmd.Run()
}

// CommitStory commits changes for a story
func (m *Manager) CommitStory(storyID, title, description string) error {
	if !m.HasChanges() {
		return nil
	}

	if err := m.StageAll(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	message := fmt.Sprintf("feat: %s\n\n%s\n\nStory: %s", title, description, storyID)
	return m.Commit(message)
}
