package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/shared/runstate"
)

func TestFileRunMetaEmbedsReviewLoopState(t *testing.T) {
	var m fileRunMeta
	if m.ReviewLoopState != (runstate.ReviewLoopState{}) {
		t.Fatal("expected zero embedded review loop state")
	}
}

func TestFileReviewLoopRoundTrip(t *testing.T) {
	workDir := t.TempDir()
	loop := NewFileReviewLoop(workDir, runstate.LocalRunID)

	if err := loop.Apply(ReviewLoopUpdate{
		Checkpoint:                 runstate.CheckpointImplReview,
		ReviewIteration:            1,
		ReviewFingerprint:          "fp1",
		ReviewElapsedMs:            100,
		LastReviewChangedFilesHash: "hash1",
	}); err != nil {
		t.Fatalf("Apply() err = %v", err)
	}

	iter, fp, elapsed, hash := loop.Snapshot()
	if iter != 1 || fp != "fp1" || elapsed != 100 || hash != "hash1" {
		t.Fatalf("Snapshot() = (%d,%q,%d,%q)", iter, fp, elapsed, hash)
	}
	if loop.Checkpoint() != runstate.CheckpointImplReview {
		t.Fatalf("Checkpoint() = %q", loop.Checkpoint())
	}

	data, err := os.ReadFile(filepath.Join(workDir, ".ralph", "runs", runstate.LocalRunID, "meta.json"))
	if err != nil {
		t.Fatalf("ReadFile meta: %v", err)
	}
	var m fileRunMeta
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if m.ReviewFingerprint != "fp1" || m.LastReviewChangedFilesHash != "hash1" {
		t.Fatalf("meta = %+v", m)
	}
}

func TestFileReviewLoopClearRecoveryAttempts(t *testing.T) {
	workDir := t.TempDir()
	loop := NewFileReviewLoop(workDir, runstate.LocalRunID)

	if err := loop.Apply(ReviewLoopUpdate{
		Checkpoint:                 runstate.CheckpointImplReview,
		ReviewIteration:            3,
		ReviewFingerprint:          "fp-abc",
		ReviewElapsedMs:            4200,
		LastReviewTranscriptPath:   "review-3.txt",
		LastReviewChangedFilesHash: "hash-xyz",
		RecoveryAttempts:           2,
	}); err != nil {
		t.Fatalf("Apply() err = %v", err)
	}
	if got := loop.RecoveryAttempts(); got != 2 {
		t.Fatalf("RecoveryAttempts() = %d, want 2", got)
	}

	if err := loop.Apply(ReviewLoopUpdate{ClearRecoveryAttempts: true}); err != nil {
		t.Fatalf("Apply(clear) err = %v", err)
	}
	if got := loop.RecoveryAttempts(); got != 0 {
		t.Fatalf("RecoveryAttempts() after clear = %d, want 0", got)
	}
	if loop.Checkpoint() != runstate.CheckpointImplReview {
		t.Fatalf("Checkpoint() = %q, want %q", loop.Checkpoint(), runstate.CheckpointImplReview)
	}
	iter, fp, elapsed, hash := loop.Snapshot()
	if iter != 3 || fp != "fp-abc" || elapsed != 4200 || hash != "hash-xyz" {
		t.Fatalf("Snapshot() = (%d,%q,%d,%q)", iter, fp, elapsed, hash)
	}
	if got := loop.LastReviewTranscriptPath(); got != "review-3.txt" {
		t.Fatalf("LastReviewTranscriptPath() = %q, want review-3.txt", got)
	}
}
