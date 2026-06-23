package runs

import (
	"fmt"
	"testing"
	"time"

	"ralph/internal/shared/runstate"
)

func TestLifecycleClarifyUpdatesStatusAndPhase(t *testing.T) {
	workDir := t.TempDir()
	reg := NewRegistry()
	run := &Run{
		ID:        "run-clarify",
		WorkDir:   workDir,
		Status:    "waiting_clarify",
		Phase:     "clarify",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	lifecycle := NewLifecycle(reg)
	if err := lifecycle.Clarify(run.ID); err != nil {
		t.Fatalf("Clarify() error = %v", err)
	}

	got, ok := reg.Get(run.ID)
	if !ok {
		t.Fatal("Get() ok = false")
	}
	if got.Status != "running" || got.Phase != "generate" {
		t.Fatalf("status/phase = %q/%q, want running/generate", got.Status, got.Phase)
	}
	if got.Checkpoint != "" {
		t.Fatalf("Checkpoint = %q, want empty", got.Checkpoint)
	}
}

func TestLifecycleTransitionsUpdateExpectedFields(t *testing.T) {
	cases := []struct {
		name       string
		apply      func(t *testing.T, l *Lifecycle, id string) error
		wantStatus string
		wantPhase  string
		wantCkpt   string
	}{
		{
			name: "review revise",
			apply: func(t *testing.T, l *Lifecycle, id string) error {
				return l.ReviseReview(id)
			},
			wantStatus: "running",
			wantPhase:  "generate",
		},
		{
			name: "implementation review continue",
			apply: func(t *testing.T, l *Lifecycle, id string) error {
				return l.ContinueImplementationReview(id)
			},
			wantStatus: "implementing",
			wantPhase:  runstate.PhaseCleanup,
		},
		{
			name: "follow up",
			apply: func(t *testing.T, l *Lifecycle, id string) error {
				return l.FollowUp(id)
			},
			wantStatus: "running",
			wantPhase:  "followup",
			wantCkpt:   runstate.CheckpointFollowup,
		},
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			workDir := t.TempDir()
			reg := NewRegistry()
			run := &Run{
				ID:        fmt.Sprintf("run-%d", i),
				WorkDir:   workDir,
				Status:    "waiting_review",
				Phase:     "review",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := reg.Register(run); err != nil {
				t.Fatalf("Register() error = %v", err)
			}

			lifecycle := NewLifecycle(reg)
			if err := tc.apply(t, lifecycle, run.ID); err != nil {
				t.Fatalf("%s() error = %v", tc.name, err)
			}

			got, ok := reg.Get(run.ID)
			if !ok {
				t.Fatal("Get() ok = false")
			}
			if got.Status != tc.wantStatus || got.Phase != tc.wantPhase {
				t.Fatalf("status/phase = %q/%q, want %q/%q", got.Status, got.Phase, tc.wantStatus, tc.wantPhase)
			}
			if got.Checkpoint != tc.wantCkpt {
				t.Fatalf("Checkpoint = %q, want %q", got.Checkpoint, tc.wantCkpt)
			}
		})
	}
}
