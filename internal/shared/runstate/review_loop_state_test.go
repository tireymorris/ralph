package runstate

import "testing"

func TestApplyReviewLoopUpdate_clearRecoveryAttemptsPreservesState(t *testing.T) {
	dst := ReviewLoopState{
		Checkpoint:                 CheckpointImplReview,
		ReviewIteration:            3,
		ReviewFingerprint:          "fp-abc",
		ReviewElapsedMs:            4200,
		StopReason:                 "guardrail",
		LastReviewTranscriptPath:   "review-3.txt",
		LastReviewChangedFilesHash: "hash-xyz",
		RecoveryAttempts:           2,
	}

	ApplyReviewLoopUpdate(&dst, ReviewLoopUpdate{ClearRecoveryAttempts: true})

	if dst.Checkpoint != CheckpointImplReview {
		t.Errorf("Checkpoint = %q, want %q", dst.Checkpoint, CheckpointImplReview)
	}
	if dst.ReviewIteration != 3 {
		t.Errorf("ReviewIteration = %d, want 3", dst.ReviewIteration)
	}
	if dst.ReviewFingerprint != "fp-abc" {
		t.Errorf("ReviewFingerprint = %q, want fp-abc", dst.ReviewFingerprint)
	}
	if dst.ReviewElapsedMs != 4200 {
		t.Errorf("ReviewElapsedMs = %d, want 4200", dst.ReviewElapsedMs)
	}
	if dst.StopReason != "guardrail" {
		t.Errorf("StopReason = %q, want guardrail", dst.StopReason)
	}
	if dst.LastReviewTranscriptPath != "review-3.txt" {
		t.Errorf("LastReviewTranscriptPath = %q, want review-3.txt", dst.LastReviewTranscriptPath)
	}
	if dst.LastReviewChangedFilesHash != "hash-xyz" {
		t.Errorf("LastReviewChangedFilesHash = %q, want hash-xyz", dst.LastReviewChangedFilesHash)
	}
	if dst.RecoveryAttempts != 0 {
		t.Errorf("RecoveryAttempts = %d, want 0", dst.RecoveryAttempts)
	}
}

func TestApplyReviewLoopUpdate_checkpointOnlyPreservesScalars(t *testing.T) {
	dst := ReviewLoopState{
		ReviewIteration:            2,
		ReviewFingerprint:          "fp",
		ReviewElapsedMs:            100,
		LastReviewChangedFilesHash: "hash",
		RecoveryAttempts:           1,
	}

	ApplyReviewLoopUpdate(&dst, ReviewLoopUpdate{Checkpoint: CheckpointPRDReview})

	if dst.Checkpoint != CheckpointPRDReview {
		t.Errorf("Checkpoint = %q, want %q", dst.Checkpoint, CheckpointPRDReview)
	}
	if dst.ReviewIteration != 2 {
		t.Errorf("ReviewIteration = %d, want 2", dst.ReviewIteration)
	}
	if dst.ReviewFingerprint != "fp" {
		t.Errorf("ReviewFingerprint = %q, want fp", dst.ReviewFingerprint)
	}
	if dst.ReviewElapsedMs != 100 {
		t.Errorf("ReviewElapsedMs = %d, want 100", dst.ReviewElapsedMs)
	}
	if dst.LastReviewChangedFilesHash != "hash" {
		t.Errorf("LastReviewChangedFilesHash = %q, want hash", dst.LastReviewChangedFilesHash)
	}
	if dst.RecoveryAttempts != 1 {
		t.Errorf("RecoveryAttempts = %d, want 1", dst.RecoveryAttempts)
	}
}

func TestApplyReviewLoopUpdate_clearsFingerprintWhenScalarBatchPresent(t *testing.T) {
	dst := ReviewLoopState{
		Checkpoint:        CheckpointImplReview,
		ReviewIteration:   2,
		ReviewFingerprint: "old-fp",
		ReviewElapsedMs:   100,
	}

	ApplyReviewLoopUpdate(&dst, ReviewLoopUpdate{
		Checkpoint:        CheckpointImplReview,
		ReviewIteration:   2,
		ReviewFingerprint: "",
		ReviewElapsedMs:   100,
	})

	if dst.ReviewFingerprint != "" {
		t.Errorf("ReviewFingerprint = %q, want empty", dst.ReviewFingerprint)
	}
}

func TestApplyReviewLoopUpdate_clearsFingerprintAtIterationZeroWithCheckpoint(t *testing.T) {
	dst := ReviewLoopState{
		Checkpoint:        CheckpointImplReview,
		ReviewFingerprint: "stale-fp",
	}
	ApplyReviewLoopUpdate(&dst, ReviewLoopUpdate{
		Checkpoint:                 CheckpointImplReview,
		ReviewFingerprint:          "",
		LastReviewChangedFilesHash: "",
	})
	if dst.ReviewFingerprint != "" {
		t.Fatalf("ReviewFingerprint = %q, want empty after clearReviewFingerprint-shaped update", dst.ReviewFingerprint)
	}
}

func TestApplyReviewLoopUpdate_clearsStopReasonOnReviewProgressUpdate(t *testing.T) {
	dst := ReviewLoopState{
		Checkpoint:        CheckpointImplReview,
		ReviewIteration:   1,
		ReviewFingerprint: "old",
		StopReason:        StopReasonDuplicateFindings,
	}
	ApplyReviewLoopUpdate(&dst, ReviewLoopUpdate{
		Checkpoint:        CheckpointImplReview,
		ReviewIteration:   2,
		ReviewFingerprint: "new-fp",
		ReviewElapsedMs:   100,
	})
	if dst.StopReason != "" {
		t.Fatalf("StopReason = %q, want empty after successful review progress update", dst.StopReason)
	}
}

func TestApplyReviewLoopUpdate_reviewProgressPreservesTranscriptPath(t *testing.T) {
	dst := ReviewLoopState{
		ReviewIteration:          1,
		ReviewFingerprint:        "old",
		LastReviewTranscriptPath: "review-1.txt",
	}
	ApplyReviewLoopUpdate(&dst, ReviewLoopUpdate{
		Checkpoint:        CheckpointImplReview,
		ReviewIteration:   2,
		ReviewFingerprint: "new-fp",
		ReviewElapsedMs:   100,
	})
	if dst.LastReviewTranscriptPath != "review-1.txt" {
		t.Fatalf("LastReviewTranscriptPath = %q, want review-1.txt preserved on progress update", dst.LastReviewTranscriptPath)
	}
}

func TestApplyReviewLoopUpdate_fullUpdateAppliesScalars(t *testing.T) {
	dst := ReviewLoopState{
		Checkpoint:       CheckpointImplReview,
		ReviewIteration:  1,
		ReviewFingerprint: "old",
	}

	ApplyReviewLoopUpdate(&dst, ReviewLoopUpdate{
		Checkpoint:                 CheckpointImplReview,
		ReviewIteration:            4,
		ReviewFingerprint:          "new-fp",
		ReviewElapsedMs:            900,
		LastReviewTranscriptPath:   "t.txt",
		LastReviewChangedFilesHash: "h",
		RecoveryAttempts:           2,
	})

	if dst.ReviewIteration != 4 {
		t.Errorf("ReviewIteration = %d, want 4", dst.ReviewIteration)
	}
	if dst.ReviewFingerprint != "new-fp" {
		t.Errorf("ReviewFingerprint = %q, want new-fp", dst.ReviewFingerprint)
	}
	if dst.ReviewElapsedMs != 900 {
		t.Errorf("ReviewElapsedMs = %d, want 900", dst.ReviewElapsedMs)
	}
	if dst.LastReviewTranscriptPath != "t.txt" {
		t.Errorf("LastReviewTranscriptPath = %q, want t.txt", dst.LastReviewTranscriptPath)
	}
	if dst.LastReviewChangedFilesHash != "h" {
		t.Errorf("LastReviewChangedFilesHash = %q, want h", dst.LastReviewChangedFilesHash)
	}
	if dst.RecoveryAttempts != 2 {
		t.Errorf("RecoveryAttempts = %d, want 2", dst.RecoveryAttempts)
	}
}
