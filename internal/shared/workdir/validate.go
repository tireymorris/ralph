package workdir

import (
	"fmt"
	"os"
	"path/filepath"
)

func ValidateGit(workDir string) error {
	info, err := os.Stat(workDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("workdir does not exist: %s", workDir)
		}
		return fmt.Errorf("workdir %s: %w", workDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("workdir is not a directory: %s", workDir)
	}
	gitPath := filepath.Join(workDir, ".git")
	gitInfo, err := os.Stat(gitPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("workdir is not a git repository: %s", workDir)
		}
		return fmt.Errorf("workdir git metadata %s: %w", gitPath, err)
	}
	if !gitInfo.IsDir() && !gitInfo.Mode().IsRegular() {
		return fmt.Errorf("workdir is not a git repository: %s", workDir)
	}
	return nil
}
