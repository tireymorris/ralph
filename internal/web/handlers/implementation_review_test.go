package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ralph/internal/web/runs"
)

func TestContinueImplementationReviewConflictWhenNotWaiting(t *testing.T) {
	api, _ := setupTestAPI(t, &runs.Run{
		ID:     "run-impl-wait",
		Prompt: "goal",
		Status: "implementing",
		Phase:  "implement",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/runs/run-impl-wait/implementation-review", nil)
	req.SetPathValue("id", "run-impl-wait")
	rec := httptest.NewRecorder()
	api.ContinueImplementationReview(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}
