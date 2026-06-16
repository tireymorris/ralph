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
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
	"ralph/internal/web/handlers"
	runctrl "ralph/internal/web/runner"
	"ralph/internal/web/runs"
	"ralph/internal/workflow/events"
)

var runIDPattern = regexp.MustCompile(`^[0-9a-f]{32}$|^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func TestCreateRunTable(t *testing.T) {
	workDir := t.TempDir()
	initGitRepoInDir(t, workDir)
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
			if tt.wantStatus == http.StatusConflict && resp["code"] != "run_conflict" {
				t.Fatalf("code = %q, want run_conflict, body = %s", resp["code"], rec.Body.String())
			}
		})
	}
}

func TestCreateRunPropagatesAutoApprove(t *testing.T) {
	workDir := t.TempDir()
	initGitRepoInDir(t, workDir)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.AutoApprove = false

	api := handlers.NewAPI(cfg, runs.NewRegistry())
	var runnerCfgAutoApprove bool
	api.SetRunnerFactory(func(c *config.Config) (runner.RunnerInterface, error) {
		runnerCfgAutoApprove = c.AutoApprove
		return &noopRunner{}, nil
	})
	t.Cleanup(api.ReleaseAllControllers)

	req := httptest.NewRequest(http.MethodPost, "/api/runs", strings.NewReader(`{"prompt":"build a feature","auto_approve":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	api.CreateRun(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if !runnerCfgAutoApprove {
		t.Fatal("runner factory config AutoApprove = false, want true")
	}
	if cfg.AutoApprove {
		t.Fatal("base API config AutoApprove was mutated")
	}
}

func TestGetRunIncludesAutoApprove(t *testing.T) {
	workDir := t.TempDir()
	initGitRepoInDir(t, workDir)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir

	api := handlers.NewAPI(cfg, runs.NewRegistry())
	api.SetRunnerFactory(func(*config.Config) (runner.RunnerInterface, error) {
		return &noopRunner{}, nil
	})
	t.Cleanup(api.ReleaseAllControllers)

	createReq := httptest.NewRequest(http.MethodPost, "/api/runs", strings.NewReader(`{"prompt":"build a feature","auto_approve":true}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	api.CreateRun(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d, body = %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	var created map[string]string
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("json.Unmarshal create: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/runs/"+created["id"], nil)
	req.SetPathValue("id", created["id"])
	rec := httptest.NewRecorder()
	api.GetRun(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body struct {
		AutoApprove bool `json:"auto_approve"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal get: %v", err)
	}
	if !body.AutoApprove {
		t.Fatalf("auto_approve = false, want true; body = %s", rec.Body.String())
	}
}

func TestGetRunUsesSharedSnapshotForStoryProgress(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	reg := runs.NewRegistry()
	now := time.Now()
	runID := "shared-snapshot-run"
	if err := reg.Register(&runs.Run{
		ID:        runID,
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
	ctrl := runctrl.NewControllerWithRunner(cfg, reg, runID, &noopRunner{})
	t.Cleanup(func() {
		ctrl.Cancel()
		ctrl.Wait()
	})
	ctrl.TrackEventState(events.EventPRDLoaded{PRD: &prd.PRD{
		ProjectName: "Shared",
		Stories: []*prd.Story{
			{
				ID:     "done",
				Title:  "Done story",
				Passes: true,
				Slices: []*prd.Slice{
					{ID: "slice-1", Behavior: "done", RedHint: "make it fail", Passes: true},
				},
			},
			{
				ID:    "active",
				Title: "Active story",
				Slices: []*prd.Slice{
					{ID: "slice-1", Behavior: "passed slice", RedHint: "make it fail", Passes: true},
					{ID: "slice-2", Behavior: "current slice", RedHint: "make it fail", RefactorHint: "extract helper", Passes: false},
				},
			},
		},
	}})
	api.SetController(runID, ctrl)

	req := httptest.NewRequest(http.MethodGet, "/api/runs/"+runID, nil)
	req.SetPathValue("id", runID)
	rec := httptest.NewRecorder()
	api.GetRun(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body struct {
		StoryProgress struct {
			Completed int `json:"completed"`
			Total     int `json:"total"`
			Stories   []struct {
				ID              string `json:"id"`
				Title           string `json:"title"`
				Passes          bool   `json:"passes"`
				CompletedSlices int    `json:"completed_slices"`
				TotalSlices     int    `json:"total_slices"`
				Slices          []struct {
					ID           string `json:"id"`
					Behavior     string `json:"behavior"`
					RedHint      string `json:"red_hint"`
					RefactorHint string `json:"refactor_hint"`
					Passes       bool   `json:"passes"`
				} `json:"slices"`
			} `json:"stories"`
		} `json:"story_progress"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.StoryProgress.Completed != 1 {
		t.Fatalf("completed = %d, want 1", body.StoryProgress.Completed)
	}
	if body.StoryProgress.Total != 2 {
		t.Fatalf("total = %d, want 2", body.StoryProgress.Total)
	}
	if len(body.StoryProgress.Stories) != 2 {
		t.Fatalf("stories = %d, want 2", len(body.StoryProgress.Stories))
	}
	active := body.StoryProgress.Stories[1]
	if active.ID != "active" || active.CompletedSlices != 1 || active.TotalSlices != 2 {
		t.Fatalf("active story = %#v, want shared snapshot counts", active)
	}
	if active.Slices[1].RefactorHint != "extract helper" {
		t.Fatalf("refactor_hint = %q, want extract helper", active.Slices[1].RefactorHint)
	}
}

func TestCreateRunValidPrompt(t *testing.T) {
	workDir := t.TempDir()
	initGitRepoInDir(t, workDir)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir

	api := handlers.NewAPI(cfg, runs.NewRegistry())
	api.SetRunnerFactory(func(*config.Config) (runner.RunnerInterface, error) {
		return &noopRunner{}, nil
	})
	t.Cleanup(api.ReleaseAllControllers)

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
    {"id": "s1", "title": "a", "description": "d", "slices": [{"id": "slice-1", "behavior": "first", "red_hint": "test it", "passes": true}], "priority": 1, "passes": true},
    {"id": "s2", "title": "b", "description": "d", "slices": [{"id": "slice-1", "behavior": "second", "red_hint": "test it", "refactor_hint": "extract helper", "passes": false}], "priority": 2, "passes": false},
    {"id": "s3", "title": "c", "description": "d", "slices": [{"id": "slice-1", "behavior": "third", "red_hint": "test it", "passes": false}], "priority": 3, "passes": false},
    {"id": "s4", "title": "d", "description": "d", "slices": [{"id": "slice-1", "behavior": "fourth", "red_hint": "test it", "passes": false}], "priority": 4, "passes": false}
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
			Stories   []struct {
				ID     string `json:"id"`
				Title  string `json:"title"`
				Passes bool   `json:"passes"`
				Slices []struct {
					ID           string `json:"id"`
					Behavior     string `json:"behavior"`
					RedHint      string `json:"red_hint"`
					RefactorHint string `json:"refactor_hint"`
					Passes       bool   `json:"passes"`
				} `json:"slices"`
			} `json:"stories"`
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
	if len(body.StoryProgress.Stories) != 4 {
		t.Fatalf("stories = %d, want 4", len(body.StoryProgress.Stories))
	}
	if body.StoryProgress.Stories[1].Slices[0].RefactorHint != "extract helper" {
		t.Fatalf("refactor_hint = %q, want extract helper", body.StoryProgress.Stories[1].Slices[0].RefactorHint)
	}
}

func TestListRunsIncludesOngoingLocalPRD(t *testing.T) {
	workDir := t.TempDir()
	prdJSON := `{
  "version": 1,
  "project_name": "CLI goal",
  "stories": [
    {"id": "s1", "title": "a", "description": "d", "acceptance_criteria": ["c"], "priority": 1, "passes": false}
  ]
}`
	if err := os.WriteFile(filepath.Join(workDir, "prd.json"), []byte(prdJSON), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	api := handlers.NewAPI(cfg, runs.NewRegistry())

	req := httptest.NewRequest(http.MethodGet, "/api/runs", nil)
	rec := httptest.NewRecorder()
	api.ListRuns(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var list []struct {
		ID     string `json:"id"`
		Prompt string `json:"prompt"`
		Source string `json:"source"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len(list) = %d, want 1", len(list))
	}
	if list[0].ID != runs.LocalPRDRunID {
		t.Fatalf("id = %q, want %q", list[0].ID, runs.LocalPRDRunID)
	}
	if list[0].Prompt != "CLI goal" {
		t.Fatalf("prompt = %q, want CLI goal", list[0].Prompt)
	}
	if list[0].Source != "local_prd" {
		t.Fatalf("source = %q, want local_prd", list[0].Source)
	}
	if list[0].Status != "implementing" {
		t.Fatalf("status = %q, want implementing", list[0].Status)
	}
}

func TestGetRunLocalPRD(t *testing.T) {
	workDir := t.TempDir()
	prdJSON := `{
  "version": 1,
  "project_name": "CLI goal",
  "stories": [
    {"id": "s1", "title": "a", "description": "d", "acceptance_criteria": ["c"], "priority": 1, "passes": false}
  ]
}`
	if err := os.WriteFile(filepath.Join(workDir, "prd.json"), []byte(prdJSON), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	api := handlers.NewAPI(cfg, runs.NewRegistry())

	req := httptest.NewRequest(http.MethodGet, "/api/runs/"+runs.LocalPRDRunID, nil)
	req.SetPathValue("id", runs.LocalPRDRunID)
	rec := httptest.NewRecorder()
	api.GetRun(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body struct {
		ID     string `json:"id"`
		Source string `json:"source"`
		Status string `json:"status"`
		Phase  string `json:"phase"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.ID != runs.LocalPRDRunID {
		t.Fatalf("id = %q, want %q", body.ID, runs.LocalPRDRunID)
	}
	if body.Source != "local_prd" {
		t.Fatalf("source = %q, want local_prd", body.Source)
	}
	if body.Status != "implementing" {
		t.Fatalf("status = %q, want implementing", body.Status)
	}
	if body.Phase != "implement" {
		t.Fatalf("phase = %q, want implement", body.Phase)
	}
}

func TestCreateRunConflictsWithLocalPRD(t *testing.T) {
	workDir := t.TempDir()
	initGitRepoInDir(t, workDir)
	prdJSON := `{
  "version": 1,
  "project_name": "CLI goal",
  "stories": [
    {"id": "s1", "title": "a", "description": "d", "acceptance_criteria": ["c"], "priority": 1, "passes": false}
  ]
}`
	if err := os.WriteFile(filepath.Join(workDir, "prd.json"), []byte(prdJSON), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	api := handlers.NewAPI(cfg, runs.NewRegistry())
	api.SetRunnerFactory(func(*config.Config) (runner.RunnerInterface, error) {
		return &noopRunner{}, nil
	})
	t.Cleanup(api.ReleaseAllControllers)

	req := httptest.NewRequest(http.MethodPost, "/api/runs", strings.NewReader(`{"prompt":"new web run"}`))
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
	if id == runs.LocalPRDRunID {
		t.Fatalf("id = %q, want a new web run id not %q", id, runs.LocalPRDRunID)
	}
	if !runIDPattern.MatchString(id) {
		t.Fatalf("id %q does not match UUID or 32+ hex", id)
	}
	if _, err := os.Stat(filepath.Join(workDir, "prd.json")); !os.IsNotExist(err) {
		t.Fatalf("prd.json should be absent at workdir root, stat err=%v", err)
	}
	backups, err := filepath.Glob(filepath.Join(workDir, ".ralph", "backups", "*"))
	if err != nil {
		t.Fatalf("glob backups: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("expected one backup dir, got %d: %v", len(backups), backups)
	}
	if _, err := os.Stat(filepath.Join(backups[0], "prd.json")); err != nil {
		t.Fatalf("archived prd.json: %v", err)
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
