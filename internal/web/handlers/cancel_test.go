package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	runctrl "ralph/internal/web/runner"
	"ralph/internal/web/runs"
	"ralph/internal/workflow/events"
)

func TestCancelRunTable(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		phase      string
		withCtrl   bool
		wantCode   int
		wantStatus string
	}{
		{"active run", "running", "clarify", true, http.StatusOK, "cancelled"},
		{"completed run", "completed", "complete", false, http.StatusConflict, "completed"},
		{"already cancelled is idempotent", "cancelled", "cancelled", false, http.StatusOK, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runID := "run-" + tt.status
			api, reg := setupTestAPI(t, &runs.Run{
				ID:      runID,
				Prompt:  "goal",
				Status:  tt.status,
				Phase:   tt.phase,
			})

			if tt.withCtrl {
				cfg := api.Cfg()
				ctrl := runctrl.NewControllerWithRunner(cfg, reg, runID, &noopRunner{})
				api.SetController(runID, ctrl)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/runs/"+runID+"/cancel", nil)
			req.SetPathValue("id", runID)
			rec := httptest.NewRecorder()
			api.CancelRun(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.wantCode, rec.Body.String())
			}
			run, ok := reg.Get(runID)
			if !ok {
				t.Fatal("run not found after cancel")
			}
			if run.Status != tt.wantStatus {
				t.Fatalf("status = %q, want %q", run.Status, tt.wantStatus)
			}
		})
	}
}

func TestCancelStopsSSELiveEventsWithin500ms(t *testing.T) {
	runID := "run-cancel-sse"
	api, reg := setupTestAPI(t, &runs.Run{
		ID:     runID,
		Prompt: "goal",
		Status: "running",
		Phase:  "clarify",
	})

	ctrl := runctrl.NewControllerWithRunner(api.Cfg(), reg, runID, &noopRunner{})
	api.SetController(runID, ctrl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var sseBody strings.Builder
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		req := httptest.NewRequest(http.MethodGet, "/api/runs/"+runID+"/events", nil).WithContext(ctx)
		req.SetPathValue("id", runID)
		rec := httptest.NewRecorder()
		api.RunEvents(rec, req)
		sseBody.WriteString(rec.Body.String())
	}()

	time.Sleep(30 * time.Millisecond)
	ctrl.EmitEvent(events.EventOutput{Output: events.Output{Text: "before-cancel"}})
	time.Sleep(30 * time.Millisecond)

	cancelReq := httptest.NewRequest(http.MethodPost, "/api/runs/"+runID+"/cancel", nil)
	cancelReq.SetPathValue("id", runID)
	cancelRec := httptest.NewRecorder()
	api.CancelRun(cancelRec, cancelReq)
	if cancelRec.Code != http.StatusOK {
		t.Fatalf("cancel status = %d, want 200", cancelRec.Code)
	}

	time.Sleep(20 * time.Millisecond)
	ctrl.EmitEvent(events.EventOutput{Output: events.Output{Text: "after-cancel-should-not-appear"}})
	time.Sleep(500 * time.Millisecond)
	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("SSE handler did not exit")
	}

	body := sseBody.String()
	if strings.Contains(body, "after-cancel-should-not-appear") {
		t.Fatalf("SSE received live event after cancel:\n%s", body)
	}
	if !strings.Contains(body, "before-cancel") {
		t.Fatalf("SSE missing pre-cancel event:\n%s", body)
	}
}
