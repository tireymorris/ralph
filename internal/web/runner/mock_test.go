package runner

import (
	"context"
	"sync"

	"ralph/internal/shared/runner"
)

type testRunner struct {
	runFunc     func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error
	runnerName  string
	commandName string
	mu          sync.Mutex
}

func (m *testRunner) Run(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
	if m.runFunc != nil {
		return m.runFunc(ctx, prompt, outputCh)
	}
	return nil
}

func (m *testRunner) RunnerName() string {
	if m.runnerName != "" {
		return m.runnerName
	}
	return "mock"
}

func (m *testRunner) CommandName() string {
	if m.commandName != "" {
		return m.commandName
	}
	return "mock-cmd"
}

func (m *testRunner) IsInternalLog(string) bool { return false }
