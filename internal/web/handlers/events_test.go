package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/web/handlers"
	runctrl "ralph/internal/web/runner"
	"ralph/internal/web/runs"
	"ralph/internal/workflow/events"
)

func TestRunEventsSSEHeaders(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	reg := runs.NewRegistry()

	api := handlers.NewAPI(cfg, reg)
	// Register a completed run so the stream closes immediately after replay.
	if err := reg.Register(&runs.Run{
		ID:      "run-sse-headers",
		WorkDir: workDir,
		Status:  "completed",
		Phase:   "complete",
	}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/runs/run-sse-headers/events", nil)
	req.SetPathValue("id", "run-sse-headers")
	rec := httptest.NewRecorder()
	api.RunEvents(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct == "" || !strings.Contains(ct, "text/event-stream") {
		t.Fatalf("Content-Type = %q, want text/event-stream", ct)
	}
}

func TestRunEventsReplayAndLive(t *testing.T) {
	workDir := t.TempDir()
	runID := "run-replay-live"
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	reg := runs.NewRegistry()
	now := time.Now()
	if err := reg.Register(&runs.Run{
		ID:        runID,
		WorkDir:   workDir,
		Prompt:    "goal",
		Status:    "running",
		Phase:     "clarify",
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	runDir := filepath.Join(workDir, ".ralph", "runs", runID)
	if err := os.MkdirAll(runDir, 0o750); err != nil {
		t.Fatal(err)
	}
	buffered := `{"type":"EventOutput","payload":{"Text":"buffered","IsErr":false,"Verbose":false}}`
	if err := os.WriteFile(filepath.Join(runDir, "events.ndjson"), []byte(buffered+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	api := handlers.NewAPI(cfg, reg)
	ctrl := runctrl.NewControllerWithRunner(cfg, reg, runID, &noopRunner{})
	api.SetController(runID, ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(20 * time.Millisecond)
		ctrl.EmitEvent(events.EventOutput{Output: events.Output{Text: "live"}})
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	req := httptest.NewRequest(http.MethodGet, "/api/runs/"+runID+"/events", nil).WithContext(ctx)
	req.SetPathValue("id", runID)
	rec := httptest.NewRecorder()
	api.RunEvents(rec, req)

	body := rec.Body.String()
	if n := strings.Count(body, "data:"); n != 2 {
		t.Fatalf("data: lines = %d, want 2, body:\n%s", n, body)
	}
}

func TestRunEventsNotFound(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	api := handlers.NewAPI(cfg, runs.NewRegistry())

	req := httptest.NewRequest(http.MethodGet, "/api/runs/unknown-id/events", nil)
	req.SetPathValue("id", "unknown-id")
	rec := httptest.NewRecorder()
	api.RunEvents(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}
