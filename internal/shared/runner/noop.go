package runner

import "context"

const defaultNoopRunnerName = "noop"

// NoopRunner is a test double that satisfies RunnerInterface without side effects.
type NoopRunner struct {
	Runner  string
	Command string
}

var _ RunnerInterface = NoopRunner{}

func (NoopRunner) Run(context.Context, string, chan<- OutputLine) error {
	return nil
}

func (r NoopRunner) RunnerName() string {
	if r.Runner != "" {
		return r.Runner
	}
	return defaultNoopRunnerName
}

func (r NoopRunner) CommandName() string {
	if r.Command != "" {
		return r.Command
	}
	return defaultNoopRunnerName
}

func (NoopRunner) IsInternalLog(string) bool {
	return false
}
