package review

import (
	"context"
	"errors"
	"testing"

	"ralph/internal/shared/runner"
)

func TestReviewDiffNonGitWorkdirReturnsGitError(t *testing.T) {
	workDir := t.TempDir()
	runner := &recordingRunner{t: t}

	_, err := ReviewDiff(context.Background(), Params{
		WorkDir:   workDir,
		RunID:     "run-1",
		Iteration: 0,
		PRDFile:   "prd.json",
		Context:   "ctx",
		Runner:    runner,
	})
	if err == nil {
		t.Fatal("ReviewDiff() err = nil, want GitError")
	}
	var gitErr *GitError
	if !errors.As(err, &gitErr) {
		t.Fatalf("ReviewDiff() err = %T %v, want *GitError", err, err)
	}
	if gitErr.WorkDir != workDir {
		t.Errorf("WorkDir = %q, want %q", gitErr.WorkDir, workDir)
	}
	if gitErr.Command == "" {
		t.Error("Command is empty, want git subcommand")
	}
	if gitErr.Output == "" {
		t.Error("Output is empty, want stderr snippet")
	}
	if runner.calls != 0 {
		t.Errorf("runner calls = %d, want 0", runner.calls)
	}
}

type recordingRunner struct {
	t     *testing.T
	calls int
}

func (r *recordingRunner) Run(context.Context, string, chan<- runner.OutputLine) error {
	r.calls++
	return nil
}

func (r *recordingRunner) RunnerName() string  { return "recording" }
func (r *recordingRunner) CommandName() string { return "recording" }
func (r *recordingRunner) IsInternalLog(string) bool {
	return false
}
