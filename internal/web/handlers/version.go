package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"ralph/internal/update"
	"ralph/internal/version"
)

var (
	versionCheck   = update.Check
	versionInstall = update.Install
)

type versionResponse struct {
	Version      string `json:"version"`
	Commit       string `json:"commit"`
	Ref          string `json:"ref"`
	Status       string `json:"status"`
	LocalCommit  string `json:"local_commit,omitempty"`
	RemoteCommit string `json:"remote_commit,omitempty"`
	CheckError   string `json:"check_error,omitempty"`
}

type updateResponse struct {
	Status       string `json:"status"`
	Message      string `json:"message"`
	Binary       string `json:"binary,omitempty"`
	LocalCommit  string `json:"local_commit,omitempty"`
	RemoteCommit string `json:"remote_commit,omitempty"`
}

func (a *API) GetVersion(w http.ResponseWriter, r *http.Request) {
	resp := versionResponse{
		Version: version.Version,
		Commit:  version.Commit,
		Ref:     version.Ref,
	}
	up, local, remote, err := versionCheck(r.Context(), update.RepoFromEnv(), update.DefaultRef)
	if err != nil {
		resp.Status = "unknown"
		resp.CheckError = err.Error()
	} else if up {
		resp.Status = "current"
		resp.LocalCommit = local
		resp.RemoteCommit = remote
	} else {
		resp.Status = "available"
		resp.LocalCommit = local
		resp.RemoteCommit = remote
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (a *API) PostUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	repo := update.RepoFromEnv()
	ref := update.DefaultRef

	up, local, remote, checkErr := versionCheck(ctx, repo, ref)
	if checkErr == nil && up {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(updateResponse{
			Status:       "current",
			Message:      "ralph is already up to date",
			LocalCommit:  local,
			RemoteCommit: remote,
		})
		return
	}
	if checkErr != nil && !strings.Contains(checkErr.Error(), "build metadata") {
		writeJSONError(w, http.StatusInternalServerError, checkErr.Error())
		return
	}

	if err := versionInstall(ctx, update.InstallOptions{Repo: repo, Ref: ref}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	bin, err := update.InstalledBinaryPath(ctx, "")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	msg := "updated; restart ralph web to use the new binary"
	if checkErr != nil {
		msg = "installed; restart ralph web to use the new binary"
		remote = ""
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(updateResponse{
		Status:       "updated",
		Message:      msg,
		Binary:       bin,
		LocalCommit:  local,
		RemoteCommit: remote,
	})
}
