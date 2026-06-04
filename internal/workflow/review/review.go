package review

import (
	"context"

	"ralph/internal/shared/runner"
)

type Params struct {
	WorkDir   string
	RunID     string
	Iteration int
	PRDFile   string
	Context   string
	Runner    runner.RunnerInterface
}

type Result struct {
	Clean                    bool
	Findings                 []Finding
	Transcript               string
	LastReviewTranscriptPath string
}

type Finding struct {
	ID       string
	Category string
	Path     string
	Line     int
	Summary  string
}

type GitError struct {
	WorkDir string
	Command string
	Output  string
}

func (e *GitError) Error() string {
	return "git error in " + e.WorkDir + ": " + e.Command + ": " + e.Output
}

func ReviewDiff(ctx context.Context, p Params) (Result, error) {
	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	default:
	}

	changed, err := branchChangedFiles(p.WorkDir)
	if err != nil {
		return Result{}, err
	}
	if len(changed) == 0 {
		return Result{Clean: true}, nil
	}

	_ = changed
	return Result{Clean: true}, nil
}
