package runner

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/web/runs"
	"ralph/internal/workflow/events"
)

func TestEventsNDJSONThreeLines(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:        "run-events",
		WorkDir:   workDir,
		Prompt:    "test",
		Status:    "running",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir

	ctrl := NewControllerWithRunner(cfg, reg, run.ID, &testRunner{})
	emit := []events.Event{
		events.EventOutput{Output: events.Output{Text: "line 1"}},
		events.EventPRDGenerating{},
		events.EventPRDGenerated{},
	}
	for _, ev := range emit {
		ctrl.EmitEvent(ev)
	}

	deadline := time.Now().Add(time.Second)
	logPath := filepath.Join(workDir, ".ralph", "runs", run.ID, "events.ndjson")
	for time.Now().Before(deadline) {
		if lines := countNDJSONLines(t, logPath); lines == 3 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("events.ndjson has %d lines, want 3", countNDJSONLines(t, logPath))
}

func countNDJSONLines(t *testing.T, path string) int {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	var n int
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		if !json.Valid(line) {
			t.Errorf("line %d is not valid JSON: %q", n+1, line)
		}
		n++
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan: %v", err)
	}
	return n
}
