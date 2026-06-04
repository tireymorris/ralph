package review

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ralph/internal/prompt"
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

	reviewPrompt := prompt.CriticalDiffReview(p.Context, p.PRDFile, changed)
	transcript, err := runRunner(ctx, p.Runner, reviewPrompt)
	if err != nil {
		return Result{}, err
	}

	relPath := transcriptRelPath(p.Iteration)
	if err := writeTranscript(p.WorkDir, p.RunID, relPath, transcript); err != nil {
		return Result{}, err
	}

	findings, err := ParseFindings(transcript)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Clean:                    len(findings) == 0,
		Findings:                 findings,
		Transcript:               transcript,
		LastReviewTranscriptPath: relPath,
	}, nil
}

func runRunner(ctx context.Context, r runner.RunnerInterface, reviewPrompt string) (string, error) {
	outputCh := make(chan runner.OutputLine, 64)
	done := make(chan struct{})
	var buf strings.Builder
	go func() {
		defer close(done)
		for line := range outputCh {
			if line.IsErr {
				buf.WriteString("STDERR: ")
			}
			buf.WriteString(line.Text)
			if !strings.HasSuffix(line.Text, "\n") {
				buf.WriteByte('\n')
			}
		}
	}()

	err := r.Run(ctx, reviewPrompt, outputCh)
	close(outputCh)
	<-done
	if err != nil {
		return buf.String(), err
	}
	return buf.String(), nil
}

func writeTranscript(workDir, runID, relPath, transcript string) error {
	dir := filepath.Join(workDir, ".ralph", "runs", runID)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create review transcript dir: %w", err)
	}
	path := filepath.Join(dir, relPath)
	if err := os.WriteFile(path, []byte(transcript), 0o600); err != nil {
		return fmt.Errorf("write review transcript %q: %w", path, err)
	}
	return nil
}
