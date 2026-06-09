package gitdiff

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type GitError struct {
	WorkDir string
	Command string
	Output  string
}

func (e *GitError) Error() string {
	return "git error in " + e.WorkDir + ": " + e.Command + ": " + e.Output
}

func ChangedFiles(workDir string) ([]string, error) {
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

func HashFiles(files []string) string {
	if len(files) == 0 {
		return ""
	}
	sorted := append([]string(nil), files...)
	sort.Strings(sorted)
	data, err := json.Marshal(sorted)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

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

func IsGitError(err error) (*GitError, bool) {
	var ge *GitError
	if errors.As(err, &ge) {
		return ge, true
	}
	return nil, false
}

func EnsureGitRepo(workDir string) error {
	if err := ensureGitRepo(workDir); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func IsUntracked(workDir, relPath string) (bool, error) {
	if err := ensureGitRepo(workDir); err != nil {
		return false, err
	}
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard", "--", relPath)
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return false, &GitError{
			WorkDir: workDir,
			Command: "git ls-files --others --exclude-standard",
			Output:  strings.TrimSpace(string(out)),
		}
	}
	return strings.TrimSpace(string(out)) == relPath, nil
}

func RemoveUntracked(workDir, relPath string) (bool, error) {
	untracked, err := IsUntracked(workDir, relPath)
	if err != nil {
		return false, err
	}
	if !untracked {
		return false, nil
	}
	path := filepath.Join(workDir, relPath)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	return true, nil
}
