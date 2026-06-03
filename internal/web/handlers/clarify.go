package handlers

import (
	"encoding/json"
	"net/http"

	"ralph/internal/prompt"
)

type clarifyRequest struct {
	Answers []prompt.QuestionAnswer `json:"answers"`
}

func (a *API) ClarifyRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	run, ok := a.registry.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "run not found")
		return
	}
	if run.Status != "waiting_clarify" {
		writeJSONError(w, http.StatusConflict, "run is not waiting for clarification")
		return
	}

	var req clarifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	a.mu.Lock()
	ctrl := a.controllers[id]
	a.mu.Unlock()
	if ctrl == nil {
		writeJSONError(w, http.StatusConflict, "run controller unavailable")
		return
	}
	if err := ctrl.SubmitClarify(req.Answers); err != nil {
		writeJSONError(w, http.StatusConflict, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}
