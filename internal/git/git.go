package git

import (
	"fmt"
	"os/exec"
	"strings"

	"ralph/internal/errors"
)

type Manager struct {
	workDir string
}

func New() *Manager {
	return &Manager{}
}

func NewWithWorkDir(workDir string) *Manager {
	return &Manager{workDir: workDir}
}

func (m *Manager) command(args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	if m.workDir != "" {
		cmd.Dir = m.workDir
	}
	return cmd
}

func (m *Manager) IsRepository() bool {
	cmd := m.command("rev-parse", "--git-dir")
	return cmd.Run() == nil
}

func (m *Manager) CurrentBranch() (string, error) {
	cmd := m.command("rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", errors.GitError{Op: "current branch", Err: err}
	}
	return strings.TrimSpace(string(out)), nil
}

func (m *Manager) BranchExists(name string) bool {
	cmd := m.command("show-ref", "--verify", "--quiet", fmt.Sprintf("refs/heads/%s", name))
	return cmd.Run() == nil
}

// If the branch already exists, it just checks it out.
func (m *Manager) CreateBranch(name string) error {
	if m.BranchExists(name) {
		if err := m.Checkout(name); err != nil {
			return errors.GitError{Op: "checkout", Err: err}
		}
		return nil
	}
	cmd := m.command("checkout", "-b", name)
	if err := cmd.Run(); err != nil {
		return errors.GitError{Op: "create branch", Err: err}
	}
	return nil
}

func (m *Manager) Checkout(name string) error {
	cmd := m.command("checkout", name)
	if err := cmd.Run(); err != nil {
		return errors.GitError{Op: "checkout", Err: err}
	}
	return nil
}

func (m *Manager) HasChanges() bool {
	cmd := m.command("diff", "--quiet", "--exit-code")
	if cmd.Run() != nil {
		return true
	}
	cmd = m.command("diff", "--staged", "--quiet", "--exit-code")
	return cmd.Run() != nil
}

func (m *Manager) StageAll() error {
	cmd := m.command("add", ".")
	if err := cmd.Run(); err != nil {
		return errors.GitError{Op: "stage all", Err: err}
	}
	return nil
}

func (m *Manager) Commit(message string) error {
	cmd := m.command("commit", "-m", message)
	if err := cmd.Run(); err != nil {
		return errors.GitError{Op: "commit", Err: err}
	}
	return nil
}

func (m *Manager) CommitStory(storyID, title, description string) error {
	if !m.HasChanges() {
		return nil
	}

	if err := m.StageAll(); err != nil {
		return errors.GitError{Op: "staging", Err: err}
	}

	message := fmt.Sprintf("feat: %s\n\n%s\n\nStory: %s", title, description, storyID)
	if err := m.Commit(message); err != nil {
		return errors.GitError{Op: "committing", Err: err}
	}
	return nil
}
