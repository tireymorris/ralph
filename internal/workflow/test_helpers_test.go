package workflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/config"
	"ralph/internal/prd"
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

func newWorkflowExecutor(t *testing.T, workDir string, eventsCh chan Event) (*Executor, *mockRunner) {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	mock := newMockRunner()
	return NewExecutorWithRunner(cfg, eventsCh, mock), mock
}

func writeQuestionsFile(t *testing.T, dir string, data string) error {
	t.Helper()
	return os.WriteFile(filepath.Join(dir, ClarifyingQuestionsFile), []byte(data), 0644)
}

func savePRD(t *testing.T, cfg *config.Config, p *prd.PRD) {
	t.Helper()
	if err := prd.Save(cfg, p); err != nil {
		t.Fatalf("failed to save test PRD: %v", err)
	}
}
