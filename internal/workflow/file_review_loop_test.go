package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/shared/runstate"
)

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
