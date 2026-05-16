package runner

import (
	"context"
	"testing"

	"ralph/internal/shared/config"
)

func TestPiRunnerRunArgs(t *testing.T) {
	cfg := &config.Config{Runner: "pi"}
	r := NewPi(cfg)

	var capturedArgs []string
	r.CmdFunc = func(ctx context.Context, name string, args ...string) CmdInterface {
		capturedArgs = args
		return &mockCmd{stdout: "", stderr: ""}
	}

	if err := r.Run(context.Background(), "test prompt", nil); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	expectedArgs := []string{"--print", "--mode", "json", "--no-session", "test prompt"}
	assertArgsEqual(t, capturedArgs, expectedArgs)
	assertNoModelSelectionArgs(t, capturedArgs)
}
