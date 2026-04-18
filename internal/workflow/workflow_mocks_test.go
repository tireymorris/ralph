package workflow

import (
	"context"
	"sync"

	"ralph/internal/runner"
)

type mockRunner struct {
	runFunc      func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error
	runnerName   string
	commandName  string
	internalLogs []string
	mu           sync.Mutex
	calls        []string
}

func newMockRunner() *mockRunner {
	return &mockRunner{
		runnerName:  "mock",
		commandName: "mock-cmd",
	}
}

func (m *mockRunner) Run(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
	m.mu.Lock()
	m.calls = append(m.calls, prompt)
	m.mu.Unlock()

	if m.runFunc != nil {
		return m.runFunc(ctx, prompt, outputCh)
	}
	return nil
}

func (m *mockRunner) RunnerName() string  { return m.runnerName }
func (m *mockRunner) CommandName() string { return m.commandName }
func (m *mockRunner) IsInternalLog(line string) bool {
	for _, l := range m.internalLogs {
		if line == l {
			return true
		}
	}
	return false
}

func (m *mockRunner) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}
