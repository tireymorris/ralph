package review

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ralph/internal/shared/runner"
)

func TestReviewDiffRespectsContextCancellation(t *testing.T) {
	workDir, _ := setupGitRepoWithWorkingTreeDiff(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runner := &recordingRunner{t: t}
	_, err := ReviewDiff(ctx, Params{
		WorkDir:   workDir,
		RunID:     "run-cancel",
		Iteration: 0,
		PRDFile:   "prd.json",
		Context:   "ctx",
		Runner:    runner,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ReviewDiff() err = %v, want context.Canceled", err)
	}
}

func TestReviewDiffInvokesRunnerOnceAndWritesTranscript(t *testing.T) {
	workDir, changedFile := setupGitRepoWithWorkingTreeDiff(t)
	runner := &recordingRunner{
		t:          t,
		transcript: "critical review transcript\n===ralph-findings===\n[]\n===/ralph-findings===\n",
	}

	result, err := ReviewDiff(context.Background(), Params{
		WorkDir:   workDir,
		RunID:     "run-diff",
		Iteration: 2,
		PRDFile:   "prd.json",
		Context:   "Go test stack",
		Runner:    runner,
	})
	if err != nil {
		t.Fatalf("ReviewDiff() err = %v", err)
	}
	if runner.calls != 1 {
		t.Fatalf("runner calls = %d, want 1", runner.calls)
	}
	if !strings.Contains(runner.lastPrompt, changedFile) {
		t.Errorf("prompt missing changed file %q:\n%s", changedFile, runner.lastPrompt)
	}
	if !strings.Contains(runner.lastPrompt, "Go test stack") {
		t.Error("prompt missing codebase context")
	}

	wantRel := "review-2.txt"
	if result.LastReviewTranscriptPath != wantRel {
		t.Errorf("LastReviewTranscriptPath = %q, want %q", result.LastReviewTranscriptPath, wantRel)
	}
	absPath := filepath.Join(workDir, ".ralph", "runs", "run-diff", wantRel)
	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("read transcript %q: %v", absPath, err)
	}
	if string(data) != runner.transcript {
		t.Errorf("transcript = %q, want %q", string(data), runner.transcript)
	}
	if !result.Clean {
		t.Error("Clean = false, want true with empty findings block")
	}
	if len(result.Findings) != 0 {
		t.Errorf("Findings = %v, want none", result.Findings)
	}
}

func TestRunRunnerConcatenatesAppendedOutput(t *testing.T) {
	t.Parallel()

	streaming := &streamingRunner{lines: []runner.OutputLine{
		{Text: "Using tool: read"},
		{Text: "===", Append: true},
		{Text: "ral", Append: true},
		{Text: "ph", Append: true},
		{Text: "-find", Append: true},
		{Text: "ings", Append: true},
		{Text: "===", Append: true},
		{Text: "\n[]\n===/ralph-findings===", Append: true},
	}}

	transcript, err := runRunner(context.Background(), streaming, "prompt")
	if err != nil {
		t.Fatalf("runRunner() err = %v", err)
	}

	want := "Using tool: read\n===ralph-findings===\n[]\n===/ralph-findings==="
	if transcript != want {
		t.Fatalf("transcript = %q, want %q", transcript, want)
	}

	findings, err := ParseFindings(transcript, true)
	if err != nil {
		t.Fatalf("ParseFindings() err = %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %v, want none", findings)
	}
}

func TestReviewDiffParsesStreamedFindingsFromTranscript(t *testing.T) {
	workDir, _ := setupGitRepoWithWorkingTreeDiff(t)
	streaming := &streamingRunner{lines: []runner.OutputLine{
		{Text: "Using tool: read"},
		{Text: "===", Append: true},
		{Text: "ral", Append: true},
		{Text: "ph", Append: true},
		{Text: "-find", Append: true},
		{Text: "ings", Append: true},
		{Text: "===", Append: true},
		{Text: "\n[]\n===/ralph-findings===", Append: true},
	}}

	result, err := ReviewDiff(context.Background(), Params{
		WorkDir:   workDir,
		RunID:     "run-streamed",
		Iteration: 1,
		PRDFile:   "prd.json",
		Context:   "ctx",
		Runner:    streaming,
	})
	if err != nil {
		t.Fatalf("ReviewDiff() err = %v", err)
	}
	if !result.Clean {
		t.Fatal("Clean = false, want true for empty streamed findings block")
	}
	if len(result.Findings) != 0 {
		t.Fatalf("Findings = %v, want none", result.Findings)
	}
}

func TestReviewDiffParsesFindingsFromTranscript(t *testing.T) {
	workDir, changedFile := setupGitRepoWithWorkingTreeDiff(t)
	transcript := `===ralph-findings===
[{"category":"bug","path":"` + changedFile + `","summary":"missing test"}]
===/ralph-findings===`
	runner := &recordingRunner{t: t, transcript: transcript}

	result, err := ReviewDiff(context.Background(), Params{
		WorkDir:   workDir,
		RunID:     "run-findings",
		Iteration: 1,
		PRDFile:   "prd.json",
		Context:   "ctx",
		Runner:    runner,
	})
	if err != nil {
		t.Fatalf("ReviewDiff() err = %v", err)
	}
	if result.Clean {
		t.Fatal("Clean = true, want false when findings present")
	}
	if len(result.Findings) != 1 {
		t.Fatalf("len(Findings) = %d, want 1", len(result.Findings))
	}
	f := result.Findings[0]
	if f.Category != "bug" || f.Path != changedFile || f.Summary != "missing test" {
		t.Errorf("Finding = %+v, want bug/%s/missing test", f, changedFile)
	}
	if f.ID == "" {
		t.Error("Finding.ID is empty")
	}
	if fp := Fingerprint(result.Findings); len(fp) != 64 {
		t.Errorf("Fingerprint() = %q, want 64-char hex", fp)
	}
}

func TestReviewDiffEmptyChangedFilesSkipsRunner(t *testing.T) {
	workDir := setupCleanGitRepo(t)
	runner := &recordingRunner{t: t}

	result, err := ReviewDiff(context.Background(), Params{
		WorkDir:   workDir,
		RunID:     "run-clean",
		Iteration: 1,
		PRDFile:   "prd.json",
		Context:   "ctx",
		Runner:    runner,
	})
	if err != nil {
		t.Fatalf("ReviewDiff() err = %v", err)
	}
	if !result.Clean {
		t.Error("Clean = false, want true")
	}
	if len(result.Findings) != 0 {
		t.Errorf("Findings = %v, want none", result.Findings)
	}
	if runner.calls != 0 {
		t.Errorf("runner calls = %d, want 0", runner.calls)
	}
	if result.LastReviewTranscriptPath != "" {
		t.Errorf("LastReviewTranscriptPath = %q, want empty", result.LastReviewTranscriptPath)
	}
}

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
	t          *testing.T
	calls      int
	lastPrompt string
	transcript string
}

func (r *recordingRunner) Run(_ context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
	r.calls++
	r.lastPrompt = prompt
	if outputCh != nil && r.transcript != "" {
		outputCh <- runner.OutputLine{Text: r.transcript}
	}
	return nil
}

func (r *recordingRunner) RunnerName() string  { return "recording" }
func (r *recordingRunner) CommandName() string { return "recording" }
func (r *recordingRunner) IsInternalLog(string) bool {
	return false
}

type streamingRunner struct {
	lines []runner.OutputLine
}

func (r *streamingRunner) Run(_ context.Context, _ string, outputCh chan<- runner.OutputLine) error {
	for _, line := range r.lines {
		outputCh <- line
	}
	return nil
}

func (r *streamingRunner) RunnerName() string  { return "streaming" }
func (r *streamingRunner) CommandName() string { return "streaming" }
func (r *streamingRunner) IsInternalLog(string) bool {
	return false
}

type blockingRunner struct {
	started chan struct{}
}

func (b *blockingRunner) Run(ctx context.Context, _ string, _ chan<- runner.OutputLine) error {
	close(b.started)
	<-ctx.Done()
	return ctx.Err()
}

func (b *blockingRunner) RunnerName() string  { return "blocking" }
func (b *blockingRunner) CommandName() string { return "blocking" }
func (b *blockingRunner) IsInternalLog(string) bool {
	return false
}

func TestReviewDiffCancelsDuringRunner(t *testing.T) {
	workDir, _ := setupGitRepoWithWorkingTreeDiff(t)
	ctx, cancel := context.WithCancel(context.Background())
	br := &blockingRunner{started: make(chan struct{})}

	errCh := make(chan error, 1)
	go func() {
		_, err := ReviewDiff(ctx, Params{
			WorkDir:   workDir,
			RunID:     "run-block",
			Iteration: 1,
			PRDFile:   "prd.json",
			Context:   "ctx",
			Runner:    br,
		})
		errCh <- err
	}()

	select {
	case <-br.started:
	case <-time.After(2 * time.Second):
		t.Fatal("runner did not start")
	}
	cancel()

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("ReviewDiff() err = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ReviewDiff did not return after cancel")
	}
}
