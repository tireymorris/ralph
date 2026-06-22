package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/runner"
	"ralph/internal/web/runs"
)

func TestReviewApproveRecreatesControllerWhenMissing(t *testing.T) {
	api, reg := setupTestAPI(t, &runs.Run{
		ID:        "run-review-restart",
		Prompt:    "goal",
		Status:    "waiting_review",
		Phase:     "review",
		PRDPath:   "prd.json",
	})

	cfg := api.Cfg()
	cfg.PRDFile = "prd.json"
	prdPath := filepath.Join(cfg.WorkDir, "prd.json")
	data := `{"version":1,"project_name":"Test","branch_name":"feature/x","stories":[{"id":"s1","title":"Story","description":"Do it","slices":[{"id":"slice-1","behavior":"AC","red_hint":"add failing test","passes":false}],"priority":1,"passes":false}]} `
	if err := os.WriteFile(prdPath, []byte(strings.TrimSpace(data)), 0644); err != nil {
		t.Fatal(err)
	}

	api.SetRunnerFactory(func(*config.Config) (runner.RunnerInterface, error) {
		return blockingImplRunner{}, nil
	})
	t.Cleanup(api.ReleaseAllControllers)

	req := httptest.NewRequest(http.MethodPost, "/api/runs/run-review-restart/review", strings.NewReader(`{"action":"approve"}`))
	req.SetPathValue("id", "run-review-restart")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	api.ReviewRun(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		got, ok := reg.Get("run-review-restart")
		if ok && strings.Contains(got.Phase, "implement") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("run did not transition to implementing")
}
