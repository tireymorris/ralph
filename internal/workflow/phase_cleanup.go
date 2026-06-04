package workflow

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"ralph/internal/prompt"
	"ralph/internal/shared/prd"
)

func (e *Executor) RunCleanup(ctx context.Context, p *prd.PRD) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	e.emit(EventCleanupStarted{})

	changedFiles := branchChangedFiles(e.cfg.WorkDir)
	cleanupPrompt := prompt.Cleanup(p.Context, e.cfg.PRDFile, changedFiles)

	if runErr := e.runWithForwardedOutput(ctx, cleanupPrompt); runErr != nil {
		e.emit(EventError{Err: fmt.Errorf("cleanup failed: %w", runErr)})
		return runErr
	}

	e.emit(EventCleanupCompleted{})
	return nil
}

func branchChangedFiles(workDir string) []string {
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

	return files
}
