package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ralph/internal/web/runs"
)

func TestGetRunIncludesReviewLoopFields(t *testing.T) {
	api, _ := setupTestAPI(t, &runs.Run{
		ID:                "run-loop",
		Prompt:            "goal",
		Status:            "implementing",
		Phase:             "implement",
		PRDPath:           "prd.json",
		Checkpoint:        runs.CheckpointImplReview,
		ReviewIteration:   2,
		ReviewFingerprint: "abc123",
		ReviewElapsedMs:   1500,
		StopReason:        "duplicate_findings",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/runs/run-loop", nil)
	req.SetPathValue("id", "run-loop")
	rec := httptest.NewRecorder()
	api.GetRun(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body struct {
		Checkpoint        string `json:"checkpoint"`
		ReviewIteration   int    `json:"review_iteration"`
		ReviewFingerprint string `json:"review_fingerprint"`
		ReviewElapsedMs   int64  `json:"review_elapsed_ms"`
		StopReason        string `json:"stop_reason"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.Checkpoint != runs.CheckpointImplReview {
		t.Fatalf("checkpoint = %q, want %q", body.Checkpoint, runs.CheckpointImplReview)
	}
	if body.ReviewIteration != 2 {
		t.Fatalf("review_iteration = %d, want 2", body.ReviewIteration)
	}
	if body.ReviewFingerprint != "abc123" {
		t.Fatalf("review_fingerprint = %q, want abc123", body.ReviewFingerprint)
	}
	if body.ReviewElapsedMs != 1500 {
		t.Fatalf("review_elapsed_ms = %d, want 1500", body.ReviewElapsedMs)
	}
	if body.StopReason != "duplicate_findings" {
		t.Fatalf("stop_reason = %q, want duplicate_findings", body.StopReason)
	}
}
