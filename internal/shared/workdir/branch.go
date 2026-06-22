package workdir

import (
	"fmt"
	"os/exec"
	"strings"
)

var fallbackDefaultBranches = []string{"main", "master", "develop", "trunk"}

func IsDefaultBranch(branchName string, defaults []string) bool {
	if branchName == "" {
		return false
	}
	names := defaults
	if len(names) == 0 {
		names = fallbackDefaultBranches
	}
	for _, name := range names {
		if branchName == name {
			return true
		}
	}
	return false
}

func DetectDefaultBranches(workDir string) []string {
	seen := make(map[string]bool)
	var out []string
	add := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		out = append(out, name)
	}

	if name, err := gitOriginDefaultBranch(workDir); err == nil {
		add(name)
	}
	for _, name := range fallbackDefaultBranches {
		add(name)
	}
	return out
}

func gitOriginDefaultBranch(workDir string) (string, error) {
	out, err := runGitCommand(workDir, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err != nil {
		return "", err
	}
	const prefix = "refs/remotes/origin/"
	if !strings.HasPrefix(out, prefix) {
		return "", fmt.Errorf("unexpected origin HEAD ref %q", out)
	}
	return strings.TrimPrefix(out, prefix), nil
}

func CurrentBranchName(workDir string) (string, error) {
	out, err := runGitCommand(workDir, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func CheckoutBranch(workDir, branchName string) error {
	_, err := runGitCommand(workDir, "checkout", "-B", branchName)
	return err
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
