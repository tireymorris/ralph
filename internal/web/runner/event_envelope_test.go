package runner

import (
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/runstate"
	"ralph/internal/web/runs"
	"ralph/internal/workflow"
	"ralph/internal/workflow/events"
)

func TestMarshalEventEnvelope_ImplementationReviewEvents(t *testing.T) {
	cases := []struct {
		name     string
		ev       events.Event
		wantType string
	}{
		{name: "started", ev: events.EventImplementationReviewStarted{Iteration: 1}, wantType: "EventImplementationReviewStarted"},
		{name: "completed", ev: events.EventImplementationReviewCompleted{Iteration: 1, Clean: true}, wantType: "EventImplementationReviewCompleted"},
		{name: "findings", ev: events.EventImplementationReview{Findings: []events.ImplementationFinding{{ID: "a", Summary: "s"}}}, wantType: "EventImplementationReview"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := MarshalEventEnvelope(tc.ev)
			if err != nil {
				t.Fatalf("MarshalEventEnvelope() error = %v", err)
			}
			var env eventEnvelope
			if err := json.Unmarshal(data, &env); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if env.Type != tc.wantType {
				t.Errorf("Type = %q, want %q", env.Type, tc.wantType)
			}
		})
	}
}

func TestMarshalEventEnvelope_RecoveryEvents(t *testing.T) {
	cases := []struct {
		name     string
		ev       events.Event
		wantType string
	}{
		{name: "started", ev: events.EventRecoveryStarted{Reason: "test_gate", Attempt: 1, Max: 2}, wantType: "EventRecoveryStarted"},
		{name: "completed", ev: events.EventRecoveryCompleted{Reason: "test_gate", Attempt: 1, Success: true}, wantType: "EventRecoveryCompleted"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := MarshalEventEnvelope(tc.ev)
			if err != nil {
				t.Fatalf("MarshalEventEnvelope() error = %v", err)
			}
			var env eventEnvelope
			if err := json.Unmarshal(data, &env); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if env.Type != tc.wantType {
				t.Errorf("Type = %q, want %q", env.Type, tc.wantType)
			}
		})
	}
}

func TestMarshalEventEnvelope_CleanupEvents(t *testing.T) {
	cases := []struct {
		name     string
		ev       events.Event
		wantType string
	}{
		{name: "started", ev: events.EventCleanupStarted{}, wantType: "EventCleanupStarted"},
		{name: "completed", ev: events.EventCleanupCompleted{}, wantType: "EventCleanupCompleted"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := MarshalEventEnvelope(tc.ev)
			if err != nil {
				t.Fatalf("MarshalEventEnvelope() error = %v", err)
			}
			var env eventEnvelope
			if err := json.Unmarshal(data, &env); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if env.Type != tc.wantType {
				t.Errorf("Type = %q, want %q", env.Type, tc.wantType)
			}
		})
	}
}

func TestMarshalEventEnvelope_SliceEvents(t *testing.T) {
	cases := []struct {
		name     string
		ev       events.Event
		wantType string
	}{
		{name: "started", ev: events.EventSliceStarted{StoryID: "story-1", SliceID: "slice-1"}, wantType: "EventSliceStarted"},
		{name: "completed", ev: events.EventSliceCompleted{StoryID: "story-1", SliceID: "slice-1"}, wantType: "EventSliceCompleted"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := MarshalEventEnvelope(tc.ev)
			if err != nil {
				t.Fatalf("MarshalEventEnvelope() error = %v", err)
			}
			var env eventEnvelope
			if err := json.Unmarshal(data, &env); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if env.Type != tc.wantType {
				t.Errorf("Type = %q, want %q", env.Type, tc.wantType)
			}
		})
	}
}

func TestMapEventToStatusPhase_CleanupEvents(t *testing.T) {
	cases := []struct {
		name string
		ev   events.Event
	}{
		{name: "started", ev: events.EventCleanupStarted{}},
		{name: "completed", ev: events.EventCleanupCompleted{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, phase := workflow.EventStatusPhase(tc.ev)
			if status != "implementing" || phase != "cleanup" {
				t.Errorf("got (%q, %q), want (%q, %q)", status, phase, "implementing", "cleanup")
			}
			if runs.IsTerminalStatus(status) {
				t.Errorf("cleanup event status %q must not be terminal", status)
			}
		})
	}
}

func TestMapEventToStatusPhase_ImplementationReviewEvents(t *testing.T) {
	cases := []struct {
		name       string
		ev         events.Event
		wantStatus string
		wantPhase  string
	}{
		{
			name:       "started",
			ev:         events.EventImplementationReviewStarted{Iteration: 1},
			wantStatus: runstate.StatusImplementing,
			wantPhase:  runstate.PhaseImplementationReview,
		},
		{
			name:       "findings",
			ev:         events.EventImplementationReview{Findings: []events.ImplementationFinding{{ID: "a", Summary: "s"}}},
			wantStatus: runstate.StatusWaitingImplReview,
			wantPhase:  runstate.PhaseImplementationReview,
		},
		{
			name:       "completed",
			ev:         events.EventImplementationReviewCompleted{Iteration: 1, Clean: true},
			wantStatus: runstate.StatusImplementing,
			wantPhase:  runstate.PhaseImplement,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, phase := workflow.EventStatusPhase(tc.ev)
			if status != tc.wantStatus || phase != tc.wantPhase {
				t.Errorf("got (%q, %q), want (%q, %q)", status, phase, tc.wantStatus, tc.wantPhase)
			}
		})
	}
}

func TestControllerHandlesCleanupEventsWithoutPanic(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:        "run-cleanup",
		WorkDir:   workDir,
		Prompt:    "test",
		Status:    "implementing",
		Phase:     "implement",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir

	ctrl := NewControllerWithRunner(cfg, reg, run.ID, &testRunner{})
	t.Cleanup(ctrl.Cancel)

	ctrl.EmitEvent(events.EventCleanupStarted{})
	ctrl.EmitEvent(events.EventCleanupCompleted{})
	ctrl.EmitEvent(events.EventCompleted{})

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		got, ok := reg.Get(run.ID)
		if ok && got.Status == "completed" && got.Phase == "complete" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	got, _ := reg.Get(run.ID)
	t.Fatalf("expected status=completed phase=complete, got status=%q phase=%q", got.Status, got.Phase)
}

func TestControllerInvokesTerminalCallbackOnce(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:        "run-terminal-once",
		WorkDir:   workDir,
		Prompt:    "test",
		Status:    "implementing",
		Phase:     "implement",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir

	ctrl := NewControllerWithRunner(cfg, reg, run.ID, &testRunner{})
	t.Cleanup(ctrl.Cancel)

	var calls int32
	ctrl.SetOnTerminal(func() {
		atomic.AddInt32(&calls, 1)
	})

	ctrl.EmitEvent(events.EventCompleted{})
	ctrl.EmitEvent(events.EventCompleted{})

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&calls) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("terminal callback calls = %d, want 1", got)
	}

	time.Sleep(200 * time.Millisecond)
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("terminal callback calls = %d after settle, want 1", got)
	}
}
