package review

import (
	"fmt"
	"os/exec"
	"strings"
)

func ensureGitRepo(workDir string) error {
	args := []string{"rev-parse", "--is-inside-work-tree"}
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &GitError{
			WorkDir: workDir,
			Command: "git " + strings.Join(args, " "),
			Output:  strings.TrimSpace(string(out)),
		}
	}
	return nil
}

func branchChangedFiles(workDir string) ([]string, error) {
	if err := ensureGitRepo(workDir); err != nil {
		return nil, err
	}

	var files []string
	seen := make(map[string]struct{})
	addFiles := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = workDir
		out, err := cmd.Output()
		if err != nil {
			return
		}
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if line == "" {
				continue
			}
			if _, ok := seen[line]; ok {
				continue
			}
			seen[line] = struct{}{}
			files = append(files, line)
		}
	}

	addFiles("diff", "--name-only", "HEAD@{upstream}...HEAD")
	addFiles("diff", "--name-only", "HEAD")
	addFiles("ls-files", "--others", "--exclude-standard")

	return files, nil
}

func transcriptRelPath(iteration int) string {
	return fmt.Sprintf("review-%d.txt", iteration)
}
