package workflow

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"ralph/internal/prompt"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
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

	outputCh := make(chan runner.OutputLine, constants.EventChannelBuffer)
	done := make(chan struct{})
	go func() {
		e.forwardOutput(outputCh)
		close(done)
	}()

	runErr := e.runner.Run(ctx, cleanupPrompt, outputCh)
	close(outputCh)
	<-done

	if runErr != nil {
		e.emit(EventError{Err: fmt.Errorf("cleanup failed: %w", runErr)})
		return runErr
	}

	e.emit(EventCleanupCompleted{})
	return nil
}

func branchChangedFiles(workDir string) []string {
	cmd := exec.Command("git", "diff", "--name-only", "HEAD@{upstream}...HEAD")
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}
