package handlers

import (
	"net/http"

	"ralph/internal/web/runs"
)

func (a *API) CancelRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	run, ok := a.registry.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "run not found")
		return
	}
	if runs.IsTerminalStatus(run.Status) && run.Status != "cancelled" {
		writeJSONError(w, http.StatusConflict, "run is already finished")
		return
	}
	if run.Status == "cancelled" {
		w.WriteHeader(http.StatusOK)
		return
	}

	a.releaseController(id)
	_ = a.registry.UpdateStatus(id, "cancelled", "cancelled")

	w.WriteHeader(http.StatusOK)
}
