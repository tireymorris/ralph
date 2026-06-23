package handlers

import (
	"context"
	"net/http"

	"ralph/internal/shared/runstate"
	"ralph/internal/shared/workdir"
)

func (a *API) ContinueImplementationReview(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	run, ok := a.registry.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "run not found")
		return
	}
	if run.Status != runstate.StatusWaitingImplReview {
		writeJSONError(w, http.StatusConflict, "run is not waiting for cleanup")
		return
	}
	if err := workdir.ValidateGit(run.WorkDir); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctrl, err := a.ensureController(id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "runner unavailable")
		return
	}
	if err := ctrl.ContinueImplementationReview(context.Background()); err != nil {
		writeJSONError(w, http.StatusConflict, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}
