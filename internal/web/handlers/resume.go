package handlers

import (
	"context"
	"net/http"

	"ralph/internal/web/runs"
)

func (a *API) ResumeRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	run, ok := a.registry.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "run not found")
		return
	}
	if runs.IsTerminalStatus(run.Status) {
		writeJSONError(w, http.StatusConflict, "run is already finished")
		return
	}

	a.releaseController(id)

	ctrl, err := a.ensureController(id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "runner unavailable")
		return
	}

	w.WriteHeader(http.StatusAccepted)

	go ctrl.ForceResume(context.Background())
}
