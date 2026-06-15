package workdir

import (
	"fmt"
	"os/exec"
	"strings"
)

func IsDefaultBranch(branchName string) bool {
	return branchName == "main" || branchName == "master"
}

func CurrentBranchName(workDir string) (string, error) {
	out, err := runGitCommand(workDir, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func runGitCommand(workDir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s in %s: %w: %s", strings.Join(args, " "), workDir, err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}
