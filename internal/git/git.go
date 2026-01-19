package git

import (
	"fmt"
	"os/exec"
	"strings"
)

type Manager struct{}

func New() *Manager {
	return &Manager{}
}

func (m *Manager) IsRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

func (m *Manager) CurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (m *Manager) BranchExists(name string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", fmt.Sprintf("refs/heads/%s", name))
	return cmd.Run() == nil
}

func (m *Manager) CreateBranch(name string) error {
	if m.BranchExists(name) {
		return m.Checkout(name)
	}
	cmd := exec.Command("git", "checkout", "-b", name)
	return cmd.Run()
}

func (m *Manager) Checkout(name string) error {
	cmd := exec.Command("git", "checkout", name)
	return cmd.Run()
}

func (m *Manager) HasChanges() bool {
	cmd := exec.Command("git", "diff", "--quiet", "--exit-code")
	if cmd.Run() != nil {
		return true
	}
	cmd = exec.Command("git", "diff", "--staged", "--quiet", "--exit-code")
	return cmd.Run() != nil
}

func (m *Manager) StageAll() error {
	cmd := exec.Command("git", "add", ".")
	return cmd.Run()
}

func (m *Manager) Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	return cmd.Run()
}

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
