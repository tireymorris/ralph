package runner

import (
	_ "embed"
	"strings"
	"testing"
	"time"

	"ralph/internal/shared/runstate"
	"ralph/internal/web/runs"
)

//go:embed review_loop.go
var reviewLoopSource string

func TestRegistryReviewLoop_readsPersistedTranscriptPath(t *testing.T) {
	reg := runs.NewRegistry()
	workDir := t.TempDir()
	runID := "run-transcript"
	run := &runs.Run{
		ID:        runID,
		WorkDir:   workDir,
		Status:    "implementing",
		Phase:     "implement",
		CreatedAt: time.Unix(0, 0).UTC(),
		UpdatedAt: time.Unix(0, 0).UTC(),
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if err := reg.UpdateReviewLoop(runID, runstate.ReviewLoopUpdate{
		Checkpoint:               runstate.CheckpointImplReview,
		LastReviewTranscriptPath: "review-1.txt",
	}); err != nil {
		t.Fatalf("UpdateReviewLoop() error = %v", err)
	}

	rl := newRegistryReviewLoop(reg, runID).(*registryReviewLoop)
	if got := rl.LastReviewTranscriptPath(); got != "review-1.txt" {
		t.Fatalf("LastReviewTranscriptPath() = %q, want review-1.txt", got)
	}
}

func TestRegistryReviewLoop_readsPersistedRecoveryAttempts(t *testing.T) {
	reg := runs.NewRegistry()
	workDir := t.TempDir()
	runID := "run-recovery"
	run := &runs.Run{
		ID:        runID,
		WorkDir:   workDir,
		Status:    "implementing",
		Phase:     "implement",
		CreatedAt: time.Unix(0, 0).UTC(),
		UpdatedAt: time.Unix(0, 0).UTC(),
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if err := reg.UpdateReviewLoop(runID, runstate.ReviewLoopUpdate{RecoveryAttempts: 2}); err != nil {
		t.Fatalf("UpdateReviewLoop() error = %v", err)
	}

	rl := newRegistryReviewLoop(reg, runID).(*registryReviewLoop)
	if got := rl.RecoveryAttempts(); got != 2 {
		t.Fatalf("RecoveryAttempts() = %d, want 2", got)
	}
}

func TestRegistryReviewLoopApplyForwardsUpdateUnmodified(t *testing.T) {
	if strings.Contains(reviewLoopSource, "Checkpoint:") {
		t.Fatal("Apply() should forward u to UpdateReviewLoop without per-field mapping")
	}
	if !strings.Contains(reviewLoopSource, "return r.registry.UpdateReviewLoop(r.runID, u)") {
		t.Fatal("Apply() should pass the update struct through to UpdateReviewLoop unchanged")
	}
}
