package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"ralph/internal/web/runs"
)

type followUpRequest struct {
	Message string `json:"message"`
}

func (a *API) FollowUpRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	run, ok := a.registry.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "run not found")
		return
	}
	if !runs.IsTerminalStatus(run.Status) {
		writeJSONError(w, http.StatusConflict, "run is not eligible for follow-up")
		return
	}

	var req followUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	message := strings.TrimSpace(req.Message)
	if message == "" {
		writeJSONError(w, http.StatusBadRequest, "message is required")
		return
	}

	ctrl, err := a.ensureController(id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "runner unavailable")
		return
	}

	w.WriteHeader(http.StatusAccepted)

	go ctrl.RunFollowUp(context.Background(), message, a.cfg)
}
