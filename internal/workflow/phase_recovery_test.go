package workflow

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/gitdiff"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runstate"
	"ralph/internal/shared/runner"
	"ralph/internal/shared/testgit"
)

func TestRunImplementationReviewRecoversFromFindings(t *testing.T) {
	workDir, _ := setupGitRepoWithWorkingTreeDiff(t)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	findingsTranscript := `===ralph-findings===
[{"category":"bug","path":"delta.txt","summary":"fix me"}]
===/ralph-findings===`

	ch := make(chan Event, 100)
	mock := newMockRunner()
	reviewCalls := 0
	mock.runFunc = func(_ context.Context, p string, outputCh chan<- runner.OutputLine) error {
		switch {
		case isDiffReviewPrompt(p):
			reviewCalls++
			if reviewCalls == 1 {
				outputCh <- runner.OutputLine{Text: findingsTranscript}
				return nil
			}
			outputCh <- runner.OutputLine{Text: cleanReviewTranscript}
		case isRecoveryPrompt(p):
			return nil
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	exec.runID = "run-recover"
	blocked, err := exec.runImplementationReview(context.Background(), &prd.PRD{Context: "ctx"})
	if err != nil {
		t.Fatalf("runImplementationReview() error = %v", err)
	}
	if blocked {
		t.Fatal("expected review to recover and pass")
	}
	if reviewCalls < 1 {
		t.Fatalf("review calls = %d, want at least 1", reviewCalls)
	}
}

func TestRunImplementationReviewRecoveryExhaustedOnDuplicateFindings(t *testing.T) {
	workDir, _ := setupGitRepoWithWorkingTreeDiff(t)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	findingsTranscript := `===ralph-findings===
[{"category":"bug","path":"delta.txt","summary":"missing test"}]
===/ralph-findings===`

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(_ context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if isDiffReviewPrompt(p) {
			outputCh <- runner.OutputLine{Text: findingsTranscript}
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	exec.runID = "run-dup"
	changed, err := gitdiff.ChangedFiles(workDir)
	if err != nil {
		t.Fatalf("ChangedFiles() err = %v", err)
	}
	exec.reviewFingerprint = reviewFingerprintFromTranscript(t, findingsTranscript)
	exec.reviewChangedFilesHash = gitdiff.HashFiles(changed)

	_, err = exec.runImplementationReview(context.Background(), &prd.PRD{Context: "ctx"})
	if err == nil {
		t.Fatal("expected recovery_exhausted error")
	}
	if !strings.Contains(err.Error(), runstate.StopReasonRecoveryExhausted) {
		t.Fatalf("error = %v, want %s", err, runstate.StopReasonRecoveryExhausted)
	}
}

func TestRunImplementationDuplicateFingerprintPersistsStopReason(t *testing.T) {
	workDir, _ := setupGitRepoWithWorkingTreeDiff(t)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	findingsTranscript := `===ralph-findings===
[{"category":"bug","path":"delta.txt","summary":"missing test"}]
===/ralph-findings===`

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(_ context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if isDiffReviewPrompt(p) {
			outputCh <- runner.OutputLine{Text: findingsTranscript}
		}
		return nil
	}

	updater := &recordingReviewLoop{}
	exec := NewExecutorWithRunner(cfg, ch, mock)
	exec.SetReviewLoop("run-stop", updater)
	changed, err := gitdiff.ChangedFiles(workDir)
	if err != nil {
		t.Fatalf("ChangedFiles() err = %v", err)
	}
	updater.fingerprint = reviewFingerprintFromTranscript(t, findingsTranscript)
	updater.changedFilesHash = gitdiff.HashFiles(changed)

	_, err = exec.runImplementationReview(context.Background(), &prd.PRD{Context: "ctx"})
	if err == nil {
		t.Fatal("expected recovery_exhausted error")
	}
	if updater.stopReason != runstate.StopReasonRecoveryExhausted {
		t.Fatalf("stop_reason = %q, want %q", updater.stopReason, runstate.StopReasonRecoveryExhausted)
	}
}

func TestRecoveryAttemptsSnapshotDoesNotReadLoopSnapshot(t *testing.T) {
	exec := NewExecutorWithRunner(config.DefaultConfig(), nil, newMockRunner())
	exec.SetReviewLoop("run-recovery", panicSnapshotRecoveryLoop{attempts: 2})

	if got := exec.recoveryAttemptsSnapshot(); got != 2 {
		t.Fatalf("recoveryAttemptsSnapshot() = %d, want 2", got)
	}
}

type panicSnapshotRecoveryLoop struct {
	attempts int
}

func (p panicSnapshotRecoveryLoop) Snapshot() (int, string, int64, string) {
	panic("Snapshot should not be called")
}

func (p panicSnapshotRecoveryLoop) Apply(ReviewLoopUpdate) error { return nil }
func (p panicSnapshotRecoveryLoop) RecoveryAttempts() int        { return p.attempts }

func TestRecoverFromReviewFailureDoesNotRetrackUntrackedFile(t *testing.T) {
	workDir := t.TempDir()
	testgit.InitRepo(t, workDir)
	helloPath := filepath.Join(workDir, "hello.txt")
	if err := os.WriteFile(helloPath, []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}
	committed, err := gitdiff.CommitChangedFiles(workDir, "ralph: story-1")
	if err != nil {
		t.Fatalf("CommitChangedFiles() err = %v", err)
	}
	if !committed {
		t.Fatal("CommitChangedFiles() committed = false, want true")
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"
	ch := make(chan Event, 100)
	executor := NewExecutorWithRunner(cfg, ch, newMockRunner())

	rmCmd := exec.Command("git", "rm", "--cached", "hello.txt")
	rmCmd.Dir = workDir
	if out, err := rmCmd.CombinedOutput(); err != nil {
		t.Fatalf("git rm --cached hello.txt: %v\n%s", err, out)
	}

	recovered, err := executor.recoverFromReviewFailure(
		context.Background(),
		&prd.PRD{Context: "ctx"},
		prompt.RecoveryReasonReviewFindings,
		"",
		[]ImplementationFinding{{
			Category: "acceptance_criteria",
			Path:     "hello.txt",
			Summary:  "file must remain untracked",
		}},
	)
	if err != nil {
		t.Fatalf("recoverFromReviewFailure() err = %v", err)
	}
	if !recovered {
		t.Fatal("recoverFromReviewFailure() recovered = false, want true")
	}

	untracked, err := gitdiff.IsUntracked(workDir, "hello.txt")
	if err != nil {
		t.Fatalf("IsUntracked() err = %v", err)
	}
	if !untracked {
		t.Fatal("hello.txt should remain untracked after Ralph recovery commit")
	}
}

func TestApplyMechanicalCleanupRemovesUntrackedArtifact(t *testing.T) {
	workDir := t.TempDir()
	testgit.InitRepo(t, workDir)
	artifact := filepath.Join(workDir, "generated.js")
	if err := os.WriteFile(artifact, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	ch := make(chan Event, 10)
	exec := NewExecutorWithRunner(cfg, ch, newMockRunner())
	exec.applyMechanicalCleanup([]ImplementationFinding{{
		Category: "wrong-target",
		Path:     "generated.js",
		Summary:  "stray generated file",
	}})

	if _, err := os.Stat(artifact); !os.IsNotExist(err) {
		t.Fatalf("expected generated.js removed, err = %v", err)
	}
}
