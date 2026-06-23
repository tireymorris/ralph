package workflow

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
	"ralph/internal/shared/testgit"
)

func TestRunCleanupContextCancelled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()
	exec := NewExecutorWithRunner(cfg, ch, mock)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := &prd.PRD{Context: "test context"}
	_, err := exec.RunCleanup(ctx, p)

	if err == nil {
		t.Fatal("RunCleanup() should return error when context is cancelled")
	}
	if err != context.Canceled {
		t.Fatalf("RunCleanup() error = %v, want context.Canceled", err)
	}

	for len(ch) > 0 {
		e := <-ch
		if _, ok := e.(EventCleanupStarted); ok {
			t.Error("EventCleanupStarted should not be emitted when context is cancelled")
		}
	}

	if mock.CallCount() != 0 {
		t.Errorf("runner should not be called, got %d calls", mock.CallCount())
	}
}

func TestRunCleanupSkipsWhenWorktreeIsClean(t *testing.T) {
	workDir := t.TempDir()
	testgit.InitRepo(t, workDir)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"
	testPRD := &prd.PRD{Context: "clean project context"}

	ch := make(chan Event, 100)
	mock := newMockRunner()
	exec := NewExecutorWithRunner(cfg, ch, mock)

	_, err := exec.RunCleanup(context.Background(), testPRD)
	if err != nil {
		t.Fatalf("RunCleanup() error = %v", err)
	}

	if mock.CallCount() != 0 {
		t.Fatalf("runner call count = %d, want 0", mock.CallCount())
	}

	evts := drainEvents(ch)
	counts := countCleanupEvents(evts)
	if counts.started != 0 {
		t.Errorf("EventCleanupStarted count = %d, want 0", counts.started)
	}

	foundSkipOutput := false
	for _, e := range evts {
		out, ok := e.(EventOutput)
		if ok && out.Text == "Skipping cleanup: no changed files" {
			foundSkipOutput = true
		}
	}
	if !foundSkipOutput {
		t.Fatalf("events = %#v, want EventOutput with cleanup skip message", evts)
	}
}

func TestRunCleanupSuccess(t *testing.T) {
	workDir, changedFile := testgit.RepoWithWorkingTreeDiff(t)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		if isDiffReviewPrompt(prompt) {
			outputCh <- runner.OutputLine{Text: cleanReviewTranscript}
			return nil
		}
		if !strings.Contains(prompt, "Match local conventions") {
			t.Error("cleanup prompt should contain style guide task list")
		}
		if !strings.Contains(prompt, "my project context") {
			t.Error("cleanup prompt should contain the PRD context")
		}
		if !strings.Contains(prompt, changedFile) {
			t.Errorf("cleanup prompt should contain changed file %q", changedFile)
		}
		outputCh <- runner.OutputLine{Text: "refactoring..."}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	p := &prd.PRD{Context: "my project context"}

	_, err := exec.RunCleanup(context.Background(), p)
	if err != nil {
		t.Fatalf("RunCleanup() error = %v", err)
	}

	if mock.CallCount() != 2 {
		t.Fatalf("runner call count = %d, want 2 (review + cleanup)", mock.CallCount())
	}

	evts := drainEvents(ch)
	counts := countCleanupEvents(evts)
	if counts.started != 1 || counts.completed != 1 {
		t.Errorf("expected 1 cleanup started and 1 completed, got started=%d completed=%d", counts.started, counts.completed)
	}

	foundOutput := false
	for _, e := range evts {
		if _, ok := e.(EventOutput); ok {
			foundOutput = true
		}
	}
	if !foundOutput {
		t.Error("expected runner output to be forwarded as EventOutput")
	}
}

func TestRunCleanupRunsAndLogsWhenChangedFilesCannotBeListed(t *testing.T) {
	if os.Getenv("RALPH_CLEANUP_LOG_HELPER") == "1" {
		runCleanupChangedFilesErrorLogHelper(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestRunCleanupRunsAndLogsWhenChangedFilesCannotBeListed")
	cmd.Env = append(os.Environ(), "RALPH_CLEANUP_LOG_HELPER=1")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start helper test: %v", err)
	}
	captured, readErr := io.ReadAll(stderr)
	if readErr != nil {
		t.Fatalf("read helper stderr: %v", readErr)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("helper test failed: %v\nstderr:\n%s", err, captured)
	}
	if !strings.Contains(string(captured), "failed to list changed files before cleanup") {
		t.Fatalf("stderr = %q, want changed-files warning", captured)
	}
}

func runCleanupChangedFilesErrorLogHelper(t *testing.T) {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()
	exec := NewExecutorWithRunner(cfg, ch, mock)

	if _, err := exec.RunCleanup(context.Background(), &prd.PRD{Context: "ctx"}); err != nil {
		t.Fatalf("RunCleanup() error = %v", err)
	}
	if mock.CallCount() != 0 {
		t.Fatalf("runner call count = %d, want 0", mock.CallCount())
	}

	evts := drainEvents(ch)
	counts := countCleanupEvents(evts)
	if counts.started != 0 || counts.completed != 0 {
		t.Fatalf("cleanup events = started %d completed %d, want 0 each", counts.started, counts.completed)
	}

	foundSkipOutput := false
	for _, e := range evts {
		out, ok := e.(EventOutput)
		if ok && out.Text == "Skipping cleanup: could not list changed files" {
			foundSkipOutput = true
		}
	}
	if !foundSkipOutput {
		t.Fatalf("events = %#v, want skip output when changed files cannot be listed", evts)
	}
}

func TestRunCleanupRunnerError(t *testing.T) {
	workDir, _ := testgit.RepoWithWorkingTreeDiff(t)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
		if isDiffReviewPrompt(prompt) {
			outputCh <- runner.OutputLine{Text: cleanReviewTranscript}
			return nil
		}
		return errors.New("something broke")
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	p := &prd.PRD{Context: "ctx"}

	_, err := exec.RunCleanup(context.Background(), p)
	if err == nil {
		t.Fatal("RunCleanup() should return error when runner fails")
	}

	if mock.CallCount() != 2 {
		t.Fatalf("runner call count = %d, want 2 (review + cleanup)", mock.CallCount())
	}

	foundError := false
	foundCompleted := false
	foundStarted := false
	for len(ch) > 0 {
		e := <-ch
		switch ev := e.(type) {
		case EventError:
			if strings.Contains(ev.Err.Error(), "cleanup") {
				foundError = true
			}
		case EventCleanupStarted:
			foundStarted = true
		case EventCleanupCompleted:
			foundCompleted = true
		}
	}

	if !foundStarted {
		t.Error("expected EventCleanupStarted before failure")
	}

	if !foundError {
		t.Error("expected EventError with message containing 'cleanup'")
	}
	if foundCompleted {
		t.Error("EventCleanupCompleted should not be emitted on runner failure")
	}
}
