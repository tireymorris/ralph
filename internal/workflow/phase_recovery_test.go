package workflow

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/gitdiff"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runstate"
	"ralph/internal/shared/runner"
)

func TestRunImplementationReviewRecoversFromFindings(t *testing.T) {
	workDir, _ := setupGitRepoWithWorkingTreeDiff(t)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	findingsTranscript := `===ralph-findings===
[{"category":"wrong-target","path":"delta.txt","summary":"fix me"}]
===/ralph-findings===`

	ch := make(chan Event, 100)
	mock := newMockRunner()
	reviewCalls := 0
	mock.runFunc = func(_ context.Context, p string, outputCh chan<- runner.OutputLine) error {
		switch {
		case strings.Contains(p, "critical diff review"):
			reviewCalls++
			if reviewCalls == 1 {
				outputCh <- runner.OutputLine{Text: findingsTranscript}
				return nil
			}
			outputCh <- runner.OutputLine{Text: cleanReviewTranscript}
		case prompt.IsRecoveryPrompt(p):
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
	if reviewCalls < 2 {
		t.Fatalf("review calls = %d, want at least 2", reviewCalls)
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
		if strings.Contains(p, "critical diff review") {
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
		if strings.Contains(p, "critical diff review") {
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

func TestApplyMechanicalCleanupRemovesUntrackedArtifact(t *testing.T) {
	workDir := t.TempDir()
	initGitRepoInDir(t, workDir)
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
