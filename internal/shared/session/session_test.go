package session

import (
	"context"
	"strings"
	"testing"

	"ralph/internal/shared/config"
)

func TestApproveReviewLoadsMissingPRDWithImplementationError(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	s := New(cfg)

	err := s.ApproveReview(context.Background(), cfg)

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "load PRD for implementation") {
		t.Fatalf("expected load PRD for implementation error, got %q", err.Error())
	}
}
