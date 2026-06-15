package workdir

import (
	"fmt"
	"os/exec"
	"strings"
)

const featureBranchPrefix = "feature/"

func IsFeatureBranch(branchName string) bool {
	return strings.HasPrefix(branchName, featureBranchPrefix)
}

func CurrentBranchName(workDir string) (string, error) {
	out, err := runGitCommand(workDir, "rev-parse", "--abbrev-ref", "HEAD")
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
