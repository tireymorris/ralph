package session

import (
	"context"
	"strings"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
)

func TestPRDForImplementationLoadsMissingPRDWithImplementationError(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	s := New(cfg)

	_, err := s.PRDForImplementation(cfg)

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "load PRD for implementation") {
		t.Fatalf("expected load PRD for implementation error, got %q", err.Error())
	}
}

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

func TestResetPRDForImplementationUnmarksAndReloadsPRD(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	original := &prd.PRD{
		ProjectName: "Reset",
		Stories: []*prd.Story{
			{ID: "done", Title: "Done", Priority: 1, Passes: true},
		},
	}
	if err := prd.Save(cfg, original); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	s := New(cfg)
	reset, err := s.ResetPRDForImplementation(cfg)
	if err != nil {
		t.Fatalf("ResetPRDForImplementation() error = %v", err)
	}
	if reset.Stories[0].Passes {
		t.Fatal("ResetPRDForImplementation() returned story with Passes=true")
	}

	reloaded, err := prd.Load(cfg)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if reloaded.Stories[0].Passes {
		t.Fatal("ResetPRDForImplementation() saved story with Passes=true")
	}
}
