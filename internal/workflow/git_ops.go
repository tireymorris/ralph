package workflow

import (
	"fmt"
	"os/exec"
	"strings"
)

func (e *Executor) saveGitState() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = e.cfg.WorkDir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git HEAD: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func (e *Executor) rollbackToState(gitState string) error {
	cmd := exec.Command("git", "reset", "--hard", gitState)
	cmd.Dir = e.cfg.WorkDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git reset failed: %w, output: %s", err, string(output))
	}
	return nil
}
