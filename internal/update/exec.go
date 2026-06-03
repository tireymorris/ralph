package update

import (
	"context"
	"fmt"
	"os/exec"
)

type CommandRunner interface {
	Output(ctx context.Context, dir, name string, args ...string) ([]byte, error)
}

type execRunner struct{}

func (execRunner) Output(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("%s %v: %w", name, args, err)
	}
	return out, nil
}

var commandRunner CommandRunner = execRunner{}
