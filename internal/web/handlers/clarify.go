package handlers

import (
	"encoding/json"
	"net/http"

	"ralph/internal/prompt"
	"ralph/internal/shared/workdir"
	"ralph/internal/web/runs"
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

	if err := workdir.ValidateGit(run.WorkDir); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req clarifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	ctrl, err := a.ensureController(id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "runner unavailable")
		return
	}
	if !ctrl.WaitingForClarify() {
		if questions, qerr := runs.LastClarifyingQuestions(run.WorkDir, id); qerr == nil && len(questions) > 0 {
			ctrl.ResumeWaitingClarify(r.Context(), run.Prompt, questions)
		}
	}
	if err := ctrl.SubmitClarify(req.Answers); err != nil {
		writeJSONError(w, http.StatusConflict, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}
