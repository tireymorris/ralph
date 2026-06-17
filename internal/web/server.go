package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"ralph/internal/shared/config"
	"ralph/internal/web/handlers"
	"ralph/internal/web/runs"
)

func NewHandler(cfg *config.Config) (http.Handler, error) {
	h, _, err := buildHandler(cfg)
	return h, err
}

func buildHandler(cfg *config.Config) (http.Handler, *handlers.API, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	registerStatic(mux)

	registry := runs.NewRegistry()
	if err := registry.LoadFromWorkDir(cfg.WorkDir); err != nil {
		return nil, nil, fmt.Errorf("load runs registry: %w", err)
	}
	api := handlers.NewAPI(cfg, registry)
	api.ReattachInterruptedRuns()
	mux.HandleFunc("POST /api/runs", api.CreateRun)
	mux.HandleFunc("GET /api/runs", api.ListRuns)
	mux.HandleFunc("GET /api/runs/{id}", api.GetRun)
	mux.HandleFunc("GET /api/runs/{id}/prd", api.GetRunPRD)
	mux.HandleFunc("GET /api/runs/{id}/events", api.RunEvents)
	mux.HandleFunc("POST /api/runs/{id}/clarify", api.ClarifyRun)
	mux.HandleFunc("POST /api/runs/{id}/review", api.ReviewRun)
	mux.HandleFunc("POST /api/runs/{id}/implementation-review", api.ContinueImplementationReview)
	mux.HandleFunc("POST /api/runs/{id}/cancel", api.CancelRun)
	mux.HandleFunc("POST /api/runs/{id}/resume", api.ResumeRun)
	mux.HandleFunc("POST /api/runs/{id}/followup", api.FollowUpRun)
	mux.HandleFunc("GET /api/version", api.GetVersion)
	mux.HandleFunc("POST /api/update", api.PostUpdate)
	mux.HandleFunc("POST /api/clean", api.CleanState)

	return mux, api, nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
