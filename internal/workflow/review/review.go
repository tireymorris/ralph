package review

import (
	"context"
	"fmt"
	"os"
	"strings"

	"ralph/internal/prompt"
	"ralph/internal/shared/gitdiff"
	"ralph/internal/shared/runner"
	"ralph/internal/shared/runpaths"
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

type GitError = gitdiff.GitError

func ReviewDiff(ctx context.Context, p Params) (Result, error) {
	changed, err := gitdiff.ChangedFiles(p.WorkDir)
	if err != nil {
		return Result{}, err
	}
	changed = gitdiff.ExcludeReviewArtifacts(changed)
	return ReviewDiffWithChanged(ctx, p, changed)
}

func ReviewDiffWithChanged(ctx context.Context, p Params, changed []string) (Result, error) {
	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	default:
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
	if err := writeTranscript(p.WorkDir, p.RunID, p.Iteration, transcript); err != nil {
		return Result{}, err
	}

	findings, err := ParseFindings(transcript, len(changed) > 0)
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
			appendTranscriptLine(&buf, line)
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

func appendTranscriptLine(buf *strings.Builder, line runner.OutputLine) {
	if line.IsErr {
		buf.WriteString("STDERR: ")
	}
	if line.Append {
		buf.WriteString(line.Text)
		return
	}
	if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n") {
		buf.WriteByte('\n')
	}
	buf.WriteString(line.Text)
	if line.Text != "" && !strings.HasSuffix(line.Text, "\n") {
		buf.WriteByte('\n')
	}
}

func transcriptRelPath(iteration int) string {
	return fmt.Sprintf("review-%d.txt", iteration)
}

func writeTranscript(workDir, runID string, iteration int, transcript string) error {
	dir := runpaths.RunDir(workDir, runID)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create review transcript dir: %w", err)
	}
	path := runpaths.ReviewTranscriptPath(workDir, runID, iteration)
	if err := os.WriteFile(path, []byte(transcript), 0o600); err != nil {
		return fmt.Errorf("write review transcript %q: %w", path, err)
	}
	return nil
}
