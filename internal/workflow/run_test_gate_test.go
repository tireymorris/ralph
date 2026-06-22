package workflow

import (
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
)

func TestRunTestGateSkipsWhenNoCommand(t *testing.T) {
	cfg := config.DefaultConfig()
	e := NewExecutor(cfg, nil)

	if err := e.runTestGate(nil); err != nil {
		t.Fatalf("runTestGate() error = %v, want nil when test command is empty", err)
	}
}

func TestRunTestGateFailsOnCommandError(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.TestCommand = "echo boom && exit 1"
	ch := make(chan Event, 10)
	e := NewExecutor(cfg, ch)

	err := e.runTestGate(&prd.PRD{})
	if err == nil {
		t.Fatal("runTestGate() error = nil, want failure")
	}

	foundErrorEvent := false
	for len(ch) > 0 {
		if ev, ok := (<-ch).(EventError); ok && ev.Err != nil {
			foundErrorEvent = true
		}
	}
	if !foundErrorEvent {
		t.Fatal("runTestGate() should emit EventError on failure")
	}
}
