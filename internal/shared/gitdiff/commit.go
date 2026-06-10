package gitdiff

import (
	"os/exec"
	"strings"
)

func ExcludeReviewArtifacts(files []string) []string {
	var out []string
	for _, f := range files {
		if isReviewArtifact(f) {
			continue
		}
		out = append(out, f)
	}
	return out
}

func isReviewArtifact(rel string) bool {
	switch {
	case rel == "prd.json.lock":
		return true
	case strings.HasPrefix(rel, ".ralph/"):
		return true
	default:
		return false
	}
}

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

func wasInHEAD(workDir, rel string) bool {
	cmd := exec.Command("git", "cat-file", "-e", "HEAD:"+rel)
	cmd.Dir = workDir
	return cmd.Run() == nil
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

// CommitRecoveryChanges commits tracked edits first, then any new untracked
// deliverable files except paths deliberately left untracked after git rm --cached.
func CommitRecoveryChanges(workDir, message string) (bool, error) {
	committed, err := CommitTrackedChanges(workDir, message)
	if err != nil || committed {
		return committed, err
	}

	files, err := ChangedFiles(workDir)
	if err != nil {
		return false, err
	}

	var toAdd []string
	for _, f := range files {
		if !shouldAutoCommit(f) {
			continue
		}
		untracked, err := IsUntracked(workDir, f)
		if err != nil || !untracked {
			continue
		}
		if wasInHEAD(workDir, f) {
			continue
		}
		toAdd = append(toAdd, f)
	}
	if len(toAdd) == 0 {
		return false, nil
	}

	addArgs := append([]string{"add", "--"}, toAdd...)
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
