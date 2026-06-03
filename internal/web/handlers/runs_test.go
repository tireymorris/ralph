package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/runner"
	"ralph/internal/web/handlers"
	"ralph/internal/web/runs"
)

var runIDPattern = regexp.MustCompile(`^[0-9a-f]{32}$|^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func TestCreateRunTable(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	reg := runs.NewRegistry()

	api := handlers.NewAPI(cfg, reg)
	api.SetRunnerFactory(func(*config.Config) (runner.RunnerInterface, error) {
		return &noopRunner{}, nil
	})

	if err := reg.Register(&runs.Run{
		ID:        "active-run",
		WorkDir:   workDir,
		Prompt:    "first",
		Status:    "running",
		Phase:     "clarify",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		PRDPath:   "prd.json",
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	tests := []struct {
		name       string
		body       string
		wantStatus int
		checkID    bool
	}{
		{
			name:       "empty prompt",
			body:       `{"prompt":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "active run conflict",
			body:       `{"prompt":"another goal"}`,
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/runs", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			api.CreateRun(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			var resp map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}
			if resp["error"] == "" {
				t.Fatalf("error field empty, body = %s", rec.Body.String())
			}
		})
	}
}

func TestCreateRunValidPrompt(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir

	api := handlers.NewAPI(cfg, runs.NewRegistry())
	api.SetRunnerFactory(func(*config.Config) (runner.RunnerInterface, error) {
		return &noopRunner{}, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/api/runs", strings.NewReader(`{"prompt":"build a feature"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	api.CreateRun(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	id := body["id"]
	if id == "" {
		t.Fatal("id field missing")
	}
	if !runIDPattern.MatchString(id) {
		t.Fatalf("id %q does not match UUID or 32+ hex", id)
	}

	metaPath := filepath.Join(workDir, ".ralph", "runs", id, "meta.json")
	if _, err := os.Stat(metaPath); err != nil {
		t.Fatalf("meta.json: %v", err)
	}
}

func TestGetRunNotFound(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	api := handlers.NewAPI(cfg, runs.NewRegistry())

	req := httptest.NewRequest(http.MethodGet, "/api/runs/unknown-id", nil)
	req.SetPathValue("id", "unknown-id")
	rec := httptest.NewRecorder()
	api.GetRun(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body["error"] == "" {
		t.Fatal("error field empty")
	}
}

func TestGetRunStoryProgress(t *testing.T) {
	workDir := t.TempDir()
	prdJSON := `{
  "version": 1,
  "project_name": "test",
  "stories": [
    {"id": "s1", "title": "a", "description": "d", "acceptance_criteria": ["c"], "priority": 1, "passes": true},
    {"id": "s2", "title": "b", "description": "d", "acceptance_criteria": ["c"], "priority": 2, "passes": false},
    {"id": "s3", "title": "c", "description": "d", "acceptance_criteria": ["c"], "priority": 3, "passes": false},
    {"id": "s4", "title": "d", "description": "d", "acceptance_criteria": ["c"], "priority": 4, "passes": false}
  ]
}`
	if err := os.WriteFile(filepath.Join(workDir, "prd.json"), []byte(prdJSON), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	reg := runs.NewRegistry()
	now := time.Now()
	if err := reg.Register(&runs.Run{
		ID:        "run-with-prd",
		WorkDir:   workDir,
		Prompt:    "goal",
		Status:    "implementing",
		Phase:     "implement",
		CreatedAt: now,
		UpdatedAt: now,
		PRDPath:   "prd.json",
	}); err != nil {
		t.Fatal(err)
	}

	api := handlers.NewAPI(cfg, reg)
	req := httptest.NewRequest(http.MethodGet, "/api/runs/run-with-prd", nil)
	req.SetPathValue("id", "run-with-prd")
	rec := httptest.NewRecorder()
	api.GetRun(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body struct {
		StoryProgress struct {
			Completed int `json:"completed"`
			Total     int `json:"total"`
		} `json:"story_progress"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.StoryProgress.Completed != 1 {
		t.Fatalf("completed = %d, want 1", body.StoryProgress.Completed)
	}
	if body.StoryProgress.Total != 4 {
		t.Fatalf("total = %d, want 4", body.StoryProgress.Total)
	}
}

func TestListRunsEmpty(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.WorkDir = t.TempDir()

	api := handlers.NewAPI(cfg, runs.NewRegistry())

	req := httptest.NewRequest(http.MethodGet, "/api/runs", nil)
	rec := httptest.NewRecorder()
	api.ListRuns(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var list []json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("len(list) = %d, want 0", len(list))
	}
}

