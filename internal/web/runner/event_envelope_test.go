package runner

import (
	"encoding/json"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/web/runs"
	"ralph/internal/workflow/events"
)

func TestMarshalEventEnvelope_CleanupStarted(t *testing.T) {
	data, err := MarshalEventEnvelope(events.EventCleanupStarted{Pass: 1, Total: 3})
	if err != nil {
		t.Fatalf("MarshalEventEnvelope() error = %v", err)
	}
	var env eventEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if env.Type != "EventCleanupStarted" {
		t.Errorf("Type = %q, want %q", env.Type, "EventCleanupStarted")
	}
	payload, err := json.Marshal(env.Payload)
	if err != nil {
		t.Fatalf("Marshal payload: %v", err)
	}
	var got struct {
		Pass  int `json:"Pass"`
		Total int `json:"Total"`
	}
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("Unmarshal payload: %v", err)
	}
	if got.Pass != 1 || got.Total != 3 {
		t.Errorf("payload Pass=%d Total=%d, want Pass=1 Total=3", got.Pass, got.Total)
	}
}

func TestMapEventToStatusPhase_CleanupStarted(t *testing.T) {
	status, phase := mapEventToStatusPhase(events.EventCleanupStarted{})
	if status != "implementing" || phase != "cleanup" {
		t.Errorf("got (%q, %q), want (%q, %q)", status, phase, "implementing", "cleanup")
	}
}

func TestMapEventToStatusPhase_CleanupCompleted(t *testing.T) {
	status, phase := mapEventToStatusPhase(events.EventCleanupCompleted{})
	if status != "implementing" || phase != "cleanup" {
		t.Errorf("got (%q, %q), want (%q, %q)", status, phase, "implementing", "cleanup")
	}
}

func TestMapEventToStatusPhase_CleanupCompletedIsNotTerminal(t *testing.T) {
	status, _ := mapEventToStatusPhase(events.EventCleanupCompleted{})
	if runs.IsTerminalStatus(status) {
		t.Errorf("EventCleanupCompleted status %q must not be terminal", status)
	}
}

func TestMarshalEventEnvelope_CleanupCompleted(t *testing.T) {
	data, err := MarshalEventEnvelope(events.EventCleanupCompleted{Pass: 2, Total: 3})
	if err != nil {
		t.Fatalf("MarshalEventEnvelope() error = %v", err)
	}
	var env eventEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if env.Type != "EventCleanupCompleted" {
		t.Errorf("Type = %q, want %q", env.Type, "EventCleanupCompleted")
	}
	payload, err := json.Marshal(env.Payload)
	if err != nil {
		t.Fatalf("Marshal payload: %v", err)
	}
	var got struct {
		Pass  int `json:"Pass"`
		Total int `json:"Total"`
	}
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("Unmarshal payload: %v", err)
	}
	if got.Pass != 2 || got.Total != 3 {
		t.Errorf("payload Pass=%d Total=%d, want Pass=2 Total=3", got.Pass, got.Total)
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
