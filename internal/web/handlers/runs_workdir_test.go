package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/runner"
	"ralph/internal/web/handlers"
	"ralph/internal/web/runs"
)

func TestCreateRunRejectsNonGitWorkdir(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	reg := runs.NewRegistry()
	api := handlers.NewAPI(cfg, reg)
	api.SetRunnerFactory(func(*config.Config) (runner.RunnerInterface, error) {
		return &noopRunner{}, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/api/runs", strings.NewReader(`{"prompt":"goal"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	api.CreateRun(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if len(reg.List()) != 0 {
		t.Fatal("run registered despite invalid workdir")
	}
}
