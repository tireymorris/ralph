package runner

import (
	_ "embed"
	"strings"
	"testing"
)

//go:embed review_loop.go
var reviewLoopSource string

func TestRegistryReviewLoopApplyForwardsUpdateUnmodified(t *testing.T) {
	if strings.Contains(reviewLoopSource, "Checkpoint:") {
		t.Fatal("Apply() should forward u to UpdateReviewLoop without per-field mapping")
	}
	if !strings.Contains(reviewLoopSource, "return r.registry.UpdateReviewLoop(r.runID, u)") {
		t.Fatal("Apply() should pass the update struct through to UpdateReviewLoop unchanged")
	}
}
