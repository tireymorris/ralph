package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"ralph/internal/shared/prd"
	"ralph/internal/web/runs"
)

type createRunRequest struct {
	Prompt string `json:"prompt"`
}

type storyProgress struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
}

type runResponse struct {
	ID            string         `json:"id"`
	Prompt        string         `json:"prompt"`
	Status        string         `json:"status"`
	Phase         string         `json:"phase"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	StoryProgress *storyProgress `json:"story_progress,omitempty"`
}

func (a *API) CreateRun(w http.ResponseWriter, r *http.Request) {
	var req createRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		writeJSONError(w, http.StatusBadRequest, "prompt is required")
		return
	}

	workDir := a.cfg.WorkDir
	if active, ok := a.registry.ActiveForWorkDir(workDir); ok {
		writeJSONError(w, http.StatusConflict, fmt.Sprintf("active run %q in progress", active.ID))
		return
	}

	id, err := newRunID()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "generate run id")
		return
	}

	now := time.Now()
	run := &runs.Run{
		ID:        id,
		WorkDir:   workDir,
		Prompt:    prompt,
		Status:    "running",
		Phase:     "clarify",
		CreatedAt: now,
		UpdatedAt: now,
		PRDPath:   a.cfg.PRDFile,
	}
	if err := a.registry.Register(run); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "register run")
		return
	}

	runner, err := a.runnerFactory(a.cfg)
	if err != nil {
		_ = a.registry.UpdateStatus(id, "failed", "failed")
		writeJSONError(w, http.StatusInternalServerError, "runner unavailable")
		return
	}

	ctrl := a.controllerFactory(a.cfg, a.registry, id, runner)
	a.registerController(id, ctrl)

	go ctrl.StartNew(context.Background(), prompt)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func newRunID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (a *API) ListRuns(w http.ResponseWriter, r *http.Request) {
	listed := a.registry.List()
	out := make([]runResponse, 0, len(listed))
	for _, run := range listed {
		out = append(out, a.runResponse(run))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(out)
}

func (a *API) GetRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	run, ok := a.registry.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "run not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(a.runResponse(run))
}

func (a *API) runResponse(run *runs.Run) runResponse {
	resp := runResponse{
		ID:        run.ID,
		Prompt:    run.Prompt,
		Status:    run.Status,
		Phase:     run.Phase,
		CreatedAt: run.CreatedAt,
		UpdatedAt: run.UpdatedAt,
	}
	if sp := a.storyProgress(run); sp != nil {
		resp.StoryProgress = sp
	}
	return resp
}

func (a *API) storyProgress(run *runs.Run) *storyProgress {
	runCfg := *a.cfg
	runCfg.WorkDir = run.WorkDir
	if run.PRDPath != "" {
		runCfg.PRDFile = run.PRDPath
	}
	prdPath := runCfg.PRDPath()
	if _, err := os.Stat(prdPath); err != nil {
		return nil
	}
	p, err := prd.Load(&runCfg)
	if err != nil {
		return nil
	}
	return &storyProgress{
		Completed: p.CompletedCount(),
		Total:     len(p.Stories),
	}
}
