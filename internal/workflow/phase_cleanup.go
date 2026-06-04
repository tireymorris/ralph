package workflow

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"ralph/internal/prompt"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/prd"
	"ralph/internal/workflow/events"
)

func (e *Executor) RunCleanup(ctx context.Context, p *prd.PRD) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	total := constants.CleanupPassCount

	for pass := 1; pass <= total; pass++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		changedFiles := branchChangedFiles(e.cfg.WorkDir)

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		progress := events.CleanupPassProgress{Pass: pass, Total: total}
		e.emit(EventCleanupStarted{CleanupPassProgress: progress})

		cleanupPrompt := prompt.Cleanup(p.Context, e.cfg.PRDFile, changedFiles, pass, total)

		if runErr := e.runWithForwardedOutput(ctx, cleanupPrompt); runErr != nil {
			e.emit(EventError{Err: fmt.Errorf("cleanup failed: %w", runErr)})
			return runErr
		}

		e.emit(EventCleanupCompleted{CleanupPassProgress: progress})
	}

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
