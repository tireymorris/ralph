package workflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/config"
	"ralph/internal/runner"
)

func newClarifyExecutor(t *testing.T, workDir string, runFunc func(context.Context, string, chan<- runner.OutputLine) error, eventsCh chan Event) *Executor {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	mock := newMockRunner()
	mock.runFunc = runFunc
	return NewExecutorWithRunner(cfg, eventsCh, mock)
}


func writeQuestionsFile(t *testing.T, dir string, data string) error {
	t.Helper()
	return os.WriteFile(filepath.Join(dir, ClarifyingQuestionsFile), []byte(data), 0644)
}

