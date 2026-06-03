package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type reviewRequest struct {
	Action   string `json:"action"`
	Critique string `json:"critique"`
}

func (a *API) ReviewRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	run, ok := a.registry.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "run not found")
		return
	}
	if run.Status != "waiting_review" {
		writeJSONError(w, http.StatusConflict, "run is not waiting for review")
		return
	}

	var req reviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	switch req.Action {
	case "approve":
	case "revise":
		if strings.TrimSpace(req.Critique) == "" {
			writeJSONError(w, http.StatusBadRequest, "critique is required for revise")
			return
		}
	default:
		writeJSONError(w, http.StatusBadRequest, "action must be approve or revise")
		return
	}

	a.mu.Lock()
	ctrl := a.controllers[id]
	a.mu.Unlock()
	if ctrl == nil {
		writeJSONError(w, http.StatusConflict, "run controller unavailable")
		return
	}

	switch req.Action {
	case "approve":
		if err := ctrl.ApproveReview(context.Background()); err != nil {
			if strings.Contains(err.Error(), "no PRD") {
				writeJSONError(w, http.StatusConflict, err.Error())
				return
			}
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
	case "revise":
		if err := ctrl.ReviseReview(context.Background(), req.Critique); err != nil {
			writeJSONError(w, http.StatusConflict, err.Error())
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
