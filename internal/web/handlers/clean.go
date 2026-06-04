package handlers

import (
	"encoding/json"
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{})
}
