package handlers

import (
	"net/http"

	"ralph/internal/clean"
)

func (a *API) CleanState(w http.ResponseWriter, r *http.Request) {
	a.ReleaseAllControllers()
	if err := clean.RemoveState(a.cfg); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.registry.Clear()
	writeJSON(w, http.StatusOK, map[string]any{})
}
