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
	"ralph/internal/shared/runner"
	"ralph/internal/shared/testgit"
	"ralph/internal/web/handlers"
	"ralph/internal/web/runs"
)

func setupTestAPI(t *testing.T, seed ...*runs.Run) (*handlers.API, *runs.Registry) {
	t.Helper()
	workDir := t.TempDir()
	testgit.InitRepo(t, workDir)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	reg := runs.NewRegistry()
	for _, r := range seed {
		if r.WorkDir == "" {
			r.WorkDir = workDir
		}
		if r.CreatedAt.IsZero() {
			r.CreatedAt = time.Now()
		}
		if r.UpdatedAt.IsZero() {
			r.UpdatedAt = time.Now()
		}
		if err := reg.Register(r); err != nil {
			t.Fatalf("Register(%s): %v", r.ID, err)
		}
	}
	api := handlers.NewAPI(cfg, reg)
	return api, reg
}

func assertPathNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("%s still exists: %v", path, err)
	}
}

func assertRegistryEmpty(t *testing.T, reg *runs.Registry) {
	t.Helper()
	if n := len(reg.List()); n != 0 {
		t.Fatalf("registry: %d runs, want 0", n)
	}
}

func assertRalphRunsRemoved(t *testing.T, workDir string) {
	t.Helper()
	assertPathNotExist(t, filepath.Join(workDir, ".ralph", "runs"))
}

func postJSON(t *testing.T, handler http.HandlerFunc, path, pathVal, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if pathVal != "" {
		req.SetPathValue("id", pathVal)
	}
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

type blockingImplRunner struct{}

func (blockingImplRunner) Run(ctx context.Context, prompt string, _ chan<- runner.OutputLine) error {
	if strings.Contains(prompt, "implementation agent") {
		<-ctx.Done()
		return ctx.Err()
	}
	return nil
}

func (blockingImplRunner) RunnerName() string        { return "mock" }
func (blockingImplRunner) CommandName() string       { return "mock" }
func (blockingImplRunner) IsInternalLog(string) bool { return false }
