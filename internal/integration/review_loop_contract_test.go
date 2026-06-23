package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ralph/internal/shared/runstate"
	"ralph/internal/web/runs"
	"ralph/internal/workflow"
)

func TestFileReviewLoopClearRecoveryPersistsReviewFieldsInMetaJSON(t *testing.T) {
	workDir := t.TempDir()
	loop := workflow.NewFileReviewLoop(workDir, runstate.LocalRunID)

	seed := runstate.ReviewLoopUpdate{
		Checkpoint:                 runstate.CheckpointImplReview,
		ReviewIteration:            3,
		ReviewFingerprint:          "fp-seed",
		ReviewElapsedMs:            4200,
		LastReviewTranscriptPath:   "review-3.txt",
		LastReviewChangedFilesHash: "hash-xyz",
		RecoveryAttempts:           2,
	}
	if err := loop.Apply(seed); err != nil {
		t.Fatalf("Apply(seed) error = %v", err)
	}
	if err := loop.Apply(runstate.ReviewLoopUpdate{ClearRecoveryAttempts: true}); err != nil {
		t.Fatalf("Apply(clear) error = %v", err)
	}

	metaPath := filepath.Join(workDir, ".ralph", "runs", runstate.LocalRunID, "meta.json")
	assertReviewLoopMetaJSON(t, metaPath, reviewLoopMetaWant{
		checkpoint:        runstate.CheckpointImplReview,
		reviewIteration:   3,
		reviewFingerprint: "fp-seed",
		reviewElapsedMs:   4200,
		transcriptPath:    "review-3.txt",
		changedFilesHash:  "hash-xyz",
		recoveryAttempts:  0,
	})
}

func TestRegistryClearRecoveryPersistsReviewFieldsInMetaJSONAndReloads(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()

	run := &runs.Run{
		ID:        "web-run-1",
		WorkDir:   workDir,
		Status:    "implementing",
		Phase:     "implement",
		CreatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	fp := "abc123def4567890abc123def4567890abc123def4567890abc123def4567890"
	if err := reg.UpdateReviewLoop("web-run-1", runstate.ReviewLoopUpdate{
		Checkpoint:                 runstate.CheckpointImplReview,
		ReviewIteration:            3,
		ReviewFingerprint:          fp,
		ReviewElapsedMs:            4200,
		StopReason:                 "guardrail",
		LastReviewTranscriptPath:   "review-3.txt",
		LastReviewChangedFilesHash: "hash-xyz",
		RecoveryAttempts:           2,
	}); err != nil {
		t.Fatalf("UpdateReviewLoop(seed) error = %v", err)
	}
	if err := reg.UpdateReviewLoop("web-run-1", runstate.ReviewLoopUpdate{
		ClearRecoveryAttempts: true,
	}); err != nil {
		t.Fatalf("UpdateReviewLoop(clear) error = %v", err)
	}

	metaPath := filepath.Join(workDir, ".ralph", "runs", "web-run-1", "meta.json")
	assertReviewLoopMetaJSON(t, metaPath, reviewLoopMetaWant{
		checkpoint:        runstate.CheckpointImplReview,
		reviewIteration:   3,
		reviewFingerprint: fp,
		reviewElapsedMs:   4200,
		transcriptPath:    "review-3.txt",
		changedFilesHash:  "hash-xyz",
		recoveryAttempts:  0,
	})

	reloaded := runs.NewRegistry()
	if err := reloaded.LoadFromWorkDir(workDir); err != nil {
		t.Fatalf("LoadFromWorkDir() error = %v", err)
	}
	got, ok := reloaded.Get("web-run-1")
	if !ok {
		t.Fatal("Get() ok = false after reload")
	}
	if got.Checkpoint != runstate.CheckpointImplReview {
		t.Errorf("Checkpoint = %q, want %q", got.Checkpoint, runstate.CheckpointImplReview)
	}
	if got.ReviewIteration != 3 {
		t.Errorf("ReviewIteration = %d, want 3", got.ReviewIteration)
	}
	if got.RecoveryAttempts != 0 {
		t.Errorf("RecoveryAttempts = %d, want 0", got.RecoveryAttempts)
	}
	if got.StopReason != "guardrail" {
		t.Errorf("StopReason = %q, want guardrail", got.StopReason)
	}
}

func TestRunMarshalsEmbeddedReviewLoopStateAtTopLevel(t *testing.T) {
	run := runs.Run{
		ID:        "id1",
		WorkDir:   "/tmp/w",
		Status:    "implementing",
		Phase:     "implement",
		CreatedAt: time.Unix(0, 0).UTC(),
		UpdatedAt: time.Unix(0, 0).UTC(),
		ReviewLoopState: runstate.ReviewLoopState{
			Checkpoint:        runstate.CheckpointImplReview,
			ReviewIteration:   2,
			ReviewFingerprint: "fp",
		},
	}

	data, err := json.Marshal(run)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	for _, key := range []string{"checkpoint", "review_iteration", "review_fingerprint", "id", "work_dir"} {
		if _, ok := m[key]; !ok {
			t.Errorf("JSON missing top-level key %q", key)
		}
	}
	if _, ok := m["ReviewLoopState"]; ok {
		t.Fatal("JSON must not nest ReviewLoopState after embedding")
	}
}

type reviewLoopMetaWant struct {
	checkpoint        string
	reviewIteration   int
	reviewFingerprint string
	reviewElapsedMs   int64
	transcriptPath    string
	changedFilesHash  string
	recoveryAttempts  int
}

func assertReviewLoopMetaJSON(t *testing.T, path string, want reviewLoopMetaWant) {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	var onDisk map[string]any
	if err := json.Unmarshal(raw, &onDisk); err != nil {
		t.Fatalf("Unmarshal meta.json error = %v", err)
	}

	assertJSONField(t, onDisk, "checkpoint", want.checkpoint)
	assertJSONField(t, onDisk, "review_iteration", float64(want.reviewIteration))
	assertJSONField(t, onDisk, "review_fingerprint", want.reviewFingerprint)
	assertJSONField(t, onDisk, "review_elapsed_ms", float64(want.reviewElapsedMs))
	assertJSONField(t, onDisk, "last_review_transcript_path", want.transcriptPath)
	assertJSONField(t, onDisk, "last_review_changed_files_hash", want.changedFilesHash)

	gotRecovery, ok := onDisk["recovery_attempts"]
	if want.recoveryAttempts == 0 {
		if ok && gotRecovery != float64(0) && gotRecovery != nil {
			t.Errorf("recovery_attempts = %v, want absent or 0", gotRecovery)
		}
		return
	}
	assertJSONField(t, onDisk, "recovery_attempts", float64(want.recoveryAttempts))
}

func assertJSONField(t *testing.T, m map[string]any, key string, want any) {
	t.Helper()
	got, ok := m[key]
	if !ok {
		t.Errorf("meta.json missing key %q", key)
		return
	}
	if got != want {
		t.Errorf("meta.json %q = %v, want %v", key, got, want)
	}
}
