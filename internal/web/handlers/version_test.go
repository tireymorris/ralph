package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/update"
	"ralph/internal/version"
	"ralph/internal/web/runs"
)

func TestGetVersionCurrent(t *testing.T) {
	const sha = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	oldCheck := versionCheck
	oldCommit := version.Commit
	defer func() {
		versionCheck = oldCheck
		version.Commit = oldCommit
	}()
	version.Commit = sha
	versionCheck = func(context.Context, string, string) (bool, string, string, error) {
		return true, sha, sha, nil
	}

	api := NewAPI(config.DefaultConfig(), runs.NewRegistry())
	rec := httptest.NewRecorder()
	api.GetVersion(rec, httptest.NewRequest(http.MethodGet, "/api/version", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var body versionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "current" {
		t.Fatalf("status = %q, want current", body.Status)
	}
}

func TestGetVersionAvailable(t *testing.T) {
	oldCheck := versionCheck
	oldCommit := version.Commit
	defer func() {
		versionCheck = oldCheck
		version.Commit = oldCommit
	}()
	version.Commit = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	versionCheck = func(context.Context, string, string) (bool, string, string, error) {
		return false, version.Commit, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", nil
	}

	api := NewAPI(config.DefaultConfig(), runs.NewRegistry())
	rec := httptest.NewRecorder()
	api.GetVersion(rec, httptest.NewRequest(http.MethodGet, "/api/version", nil))

	var body versionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "available" {
		t.Fatalf("status = %q, want available", body.Status)
	}
}

func TestPostUpdateAlreadyCurrent(t *testing.T) {
	oldCheck := versionCheck
	oldInstall := versionInstall
	defer func() {
		versionCheck = oldCheck
		versionInstall = oldInstall
	}()
	versionCheck = func(context.Context, string, string) (bool, string, string, error) {
		return true, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil
	}
	versionInstall = func(context.Context, update.InstallOptions) error {
		t.Fatal("Install should not run")
		return nil
	}

	api := NewAPI(config.DefaultConfig(), runs.NewRegistry())
	rec := httptest.NewRecorder()
	api.PostUpdate(rec, httptest.NewRequest(http.MethodPost, "/api/update", nil))

	var body updateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "current" {
		t.Fatalf("status = %q, want current", body.Status)
	}
}

func TestPostUpdateInstalls(t *testing.T) {
	oldCheck := versionCheck
	oldInstall := versionInstall
	defer func() {
		versionCheck = oldCheck
		versionInstall = oldInstall
	}()
	versionCheck = func(context.Context, string, string) (bool, string, string, error) {
		return false, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", nil
	}
	versionInstall = func(context.Context, update.InstallOptions) error {
		return nil
	}

	api := NewAPI(config.DefaultConfig(), runs.NewRegistry())
	rec := httptest.NewRecorder()
	api.PostUpdate(rec, httptest.NewRequest(http.MethodPost, "/api/update", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var body updateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "updated" {
		t.Fatalf("status = %q, want updated", body.Status)
	}
}

func TestPostUpdateInstallFailure(t *testing.T) {
	oldCheck := versionCheck
	oldInstall := versionInstall
	defer func() {
		versionCheck = oldCheck
		versionInstall = oldInstall
	}()
	versionCheck = func(context.Context, string, string) (bool, string, string, error) {
		return false, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", nil
	}
	versionInstall = func(context.Context, update.InstallOptions) error {
		return errors.New("install failed")
	}

	api := NewAPI(config.DefaultConfig(), runs.NewRegistry())
	rec := httptest.NewRecorder()
	api.PostUpdate(rec, httptest.NewRequest(http.MethodPost, "/api/update", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}
