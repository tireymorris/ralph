package gitdiff

import (
	"os/exec"
	"strings"
)

func shouldAutoCommit(rel string) bool {
	switch {
	case rel == "prd.json", rel == "prd.json.lock":
		return false
	case strings.HasPrefix(rel, ".ralph/"):
		return false
	case strings.HasPrefix(rel, "tmp/"):
		return false
	default:
		return true
	}
}

// CommitTrackedChanges stages and commits only changes to files already known to
// git. New untracked files are left alone so recovery fixes like git rm --cached
// are not undone by a follow-up git add.
func CommitTrackedChanges(workDir, message string) (bool, error) {
	if err := ensureGitRepo(workDir); err != nil {
		return false, err
	}

	addCmd := exec.Command("git", "add", "-u")
	addCmd.Dir = workDir
	if out, err := addCmd.CombinedOutput(); err != nil {
		return false, &GitError{
			WorkDir: workDir,
			Command: "git add -u",
			Output:  strings.TrimSpace(string(out)),
		}
	}

	for _, rel := range []string{"prd.json", "prd.json.lock"} {
		reset := exec.Command("git", "reset", "HEAD", "--", rel)
		reset.Dir = workDir
		_, _ = reset.CombinedOutput()
	}

	diffCmd := exec.Command("git", "diff", "--cached", "--quiet")
	diffCmd.Dir = workDir
	if err := diffCmd.Run(); err == nil {
		return false, nil
	}

	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = workDir
	out, err := commitCmd.CombinedOutput()
	if err != nil {
		output := strings.TrimSpace(string(out))
		if strings.Contains(output, "nothing to commit") {
			return false, nil
		}
		return false, &GitError{
			WorkDir: workDir,
			Command: "git commit",
			Output:  output,
		}
	}

	return true, nil
}

func CommitChangedFiles(workDir, message string) (bool, error) {
	if err := ensureGitRepo(workDir); err != nil {
		return false, err
	}

	files, err := ChangedFiles(workDir)
	if err != nil {
		return false, err
	}

	var toCommit []string
	for _, f := range files {
		if shouldAutoCommit(f) {
			toCommit = append(toCommit, f)
		}
	}
	if len(toCommit) == 0 {
		return false, nil
	}

	addArgs := append([]string{"add", "--"}, toCommit...)
	addCmd := exec.Command("git", addArgs...)
	addCmd.Dir = workDir
	if out, err := addCmd.CombinedOutput(); err != nil {
		return false, &GitError{
			WorkDir: workDir,
			Command: "git " + strings.Join(addArgs, " "),
			Output:  strings.TrimSpace(string(out)),
		}
	}

	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = workDir
	out, err := commitCmd.CombinedOutput()
	if err != nil {
		output := strings.TrimSpace(string(out))
		if strings.Contains(output, "nothing to commit") {
			return false, nil
		}
		return false, &GitError{
			WorkDir: workDir,
			Command: "git commit",
			Output:  output,
		}
	}

	return true, nil
}
