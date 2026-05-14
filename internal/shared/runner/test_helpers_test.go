package runner

import (
	"context"
	"testing"

	"ralph/internal/shared/config"
)

func newTestRunner(t *testing.T, cfg *config.Config) *Runner {
	t.Helper()
	return &Runner{cfg: cfg, CmdFunc: defaultCmdFunc(cfg.WorkDir)}
}

func stubCmdFunc(mock CmdInterface, capturedName *string, capturedArgs *[]string) func(context.Context, string, ...string) CmdInterface {
	return func(ctx context.Context, name string, args ...string) CmdInterface {
		if capturedName != nil {
			*capturedName = name
		}
		if capturedArgs != nil {
			*capturedArgs = append((*capturedArgs)[:0], args...)
		}
		return mock
	}
}

func assertRunnerIs[T any](t *testing.T, got any) T {
	t.Helper()
	r, ok := got.(T)
	if !ok {
		t.Fatalf("got %T", got)
	}
	return r
}
