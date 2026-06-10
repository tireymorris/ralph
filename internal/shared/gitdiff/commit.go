package gitdiff

import (
	"os/exec"
	"strings"
)

func shouldAutoCommit(rel string) bool {
	switch {
	case rel == "prd.json.lock":
		return false
	case strings.HasPrefix(rel, ".ralph/"):
		return false
	case strings.HasPrefix(rel, "tmp/"):
		return false
	default:
		return true
	}
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
