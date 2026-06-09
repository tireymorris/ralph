package runner

import (
	"context"
	"io"
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

func assertArgsEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("args = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("args = %v, want %v", got, want)
		}
	}
}

func assertNoModelSelectionArgs(t *testing.T, args []string) {
	t.Helper()
	for _, arg := range args {
		if arg == "--model" || arg == "--runner" || arg == "--provider" {
			t.Fatalf("unexpected model selection arg %q in %v", arg, args)
		}
	}
}

func assertPromptDeliveredViaStdin(t *testing.T, mock *mockCmd, want string) {
	t.Helper()
	if mock.stdin == nil {
		t.Fatal("prompt must be delivered via stdin")
	}
	got, err := io.ReadAll(mock.stdin)
	if err != nil {
		t.Fatalf("ReadAll(stdin) error = %v", err)
	}
	if string(got) != want {
		t.Fatalf("stdin prompt = %d bytes, want %d bytes", len(got), len(want))
	}
}
