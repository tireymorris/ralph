package handlers

import (
	"net/http"

	"ralph/internal/clean"
)

func (a *API) CleanState(w http.ResponseWriter, r *http.Request) {
	a.ReleaseAllControllers()
	if err := clean.RemoveState(a.cfg); err != nil {
		writeJSONError(w, http.StatusInternalServerError,
			"clean failed after stopping active runs; retry clean or check file permissions: "+err.Error())
		return
	}
	a.registry.Clear()
	writeJSON(w, http.StatusOK, map[string]any{})
}
