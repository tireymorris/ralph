package app

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"ralph/internal/update"
)

func TestRunUpdateSuccess(t *testing.T) {
	oldInstall := updateInstall
	oldCheck := updateCheck
	defer func() {
		updateInstall = oldInstall
		updateCheck = oldCheck
	}()
	updateCheck = func(context.Context, string, string) (bool, string, string, error) {
		return false, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", nil
	}
	installCalled := false
	updateInstall = func(ctx context.Context, opts update.InstallOptions) error {
		installCalled = true
		return nil
	}

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	codeCh := make(chan int, 1)
	go func() {
		codeCh <- Run([]string{"update"})
		w.Close()
	}()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	os.Stdout = oldStdout

	if code := <-codeCh; code != 0 {
		t.Errorf("Run(update) = %d, want 0", code)
	}
	out := buf.String()
	if !installCalled {
		t.Fatal("Install was not called")
	}
	if !strings.Contains(out, "updating ralph") {
		t.Errorf("stdout = %q, want substring updating ralph", out)
	}
	if !strings.Contains(out, "updated:") {
		t.Errorf("stdout = %q, want substring updated:", out)
	}
	if strings.Contains(out, "already up to date") {
		t.Errorf("stdout = %q, should not say already up to date", out)
	}
}

func TestRunUpdateAlreadyUpToDate(t *testing.T) {
	const sha = "cccccccccccccccccccccccccccccccccccccccc"
	oldInstall := updateInstall
	oldCheck := updateCheck
	defer func() {
		updateInstall = oldInstall
		updateCheck = oldCheck
	}()
	updateCheck = func(context.Context, string, string) (bool, string, string, error) {
		return true, sha, sha, nil
	}
	updateInstall = func(context.Context, update.InstallOptions) error {
		t.Fatal("Install should not run when already up to date")
		return nil
	}

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	code := Run([]string{"update"})
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	if code != 0 {
		t.Errorf("Run(update) = %d, want 0", code)
	}
	out := buf.String()
	if !strings.Contains(out, "already up to date") {
		t.Errorf("stdout = %q, want substring already up to date", out)
	}
	if !strings.Contains(out, sha) {
		t.Errorf("stdout = %q, want commit %s", out, sha)
	}
	if strings.Contains(out, "updating ralph") || strings.Contains(out, "updated:") {
		t.Errorf("stdout = %q, should not install or report update", out)
	}
}

func TestRunUpdateFailure(t *testing.T) {
	oldInstall := updateInstall
	oldCheck := updateCheck
	defer func() {
		updateInstall = oldInstall
		updateCheck = oldCheck
	}()
	updateCheck = func(context.Context, string, string) (bool, string, string, error) {
		return false, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", nil
	}
	updateInstall = func(ctx context.Context, opts update.InstallOptions) error {
		return context.Canceled
	}

	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	code := Run([]string{"update"})
	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	if code != 1 {
		t.Errorf("Run(update) = %d, want 1", code)
	}
	if !strings.Contains(buf.String(), "Error") {
		t.Errorf("stderr = %q, want substring Error", buf.String())
	}
}

func TestMaybePromptedUpdateUpToDate(t *testing.T) {
	oldCheck := updateCheck
	oldInstall := updateInstall
	oldPrompt := promptUpdateConfirm
	defer func() {
		updateCheck = oldCheck
		updateInstall = oldInstall
		promptUpdateConfirm = oldPrompt
	}()

	updateCheck = func(context.Context, string, string) (bool, string, string, error) {
		return true, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil
	}
	updateInstall = func(context.Context, update.InstallOptions) error {
		t.Fatal("Install should not run when up to date")
		return nil
	}
	promptUpdateConfirm = func(string, string, bool) bool {
		t.Fatal("Prompt should not run when up to date")
		return false
	}

	if err := maybePromptedUpdate(context.Background(), "repo", "main", true); err != nil {
		t.Fatalf("maybePromptedUpdate() = %v, want nil", err)
	}
}

func TestMaybePromptedUpdateNonInteractiveSkipsInstall(t *testing.T) {
	oldCheck := updateCheck
	oldInstall := updateInstall
	oldPrompt := promptUpdateConfirm
	defer func() {
		updateCheck = oldCheck
		updateInstall = oldInstall
		promptUpdateConfirm = oldPrompt
	}()

	updateCheck = func(context.Context, string, string) (bool, string, string, error) {
		return false, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", nil
	}
	updateInstall = func(context.Context, update.InstallOptions) error {
		t.Fatal("Install should not run when non-interactive")
		return nil
	}
	promptUpdateConfirm = func(string, string, bool) bool {
		t.Fatal("Prompt should not run when non-interactive")
		return false
	}

	if err := maybePromptedUpdate(context.Background(), "repo", "main", false); err != nil {
		t.Fatalf("maybePromptedUpdate() = %v, want nil", err)
	}
}

func TestMaybePromptedUpdateDeclined(t *testing.T) {
	oldCheck := updateCheck
	oldInstall := updateInstall
	oldPrompt := promptUpdateConfirm
	defer func() {
		updateCheck = oldCheck
		updateInstall = oldInstall
		promptUpdateConfirm = oldPrompt
	}()

	updateCheck = func(context.Context, string, string) (bool, string, string, error) {
		return false, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", nil
	}
	updateInstall = func(context.Context, update.InstallOptions) error {
		t.Fatal("Install should not run when update declined")
		return nil
	}
	promptUpdateConfirm = func(string, string, bool) bool { return false }

	if err := maybePromptedUpdate(context.Background(), "repo", "main", true); err != nil {
		t.Fatalf("maybePromptedUpdate() = %v, want nil", err)
	}
}

func TestMaybePromptedUpdateAccepted(t *testing.T) {
	oldCheck := updateCheck
	oldInstall := updateInstall
	oldPrompt := promptUpdateConfirm
	defer func() {
		updateCheck = oldCheck
		updateInstall = oldInstall
		promptUpdateConfirm = oldPrompt
	}()

	updateCheck = func(context.Context, string, string) (bool, string, string, error) {
		return false, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", nil
	}
	installCalled := false
	updateInstall = func(context.Context, update.InstallOptions) error {
		installCalled = true
		return nil
	}
	promptUpdateConfirm = func(string, string, bool) bool { return true }

	if err := maybePromptedUpdate(context.Background(), "repo", "main", true); err != nil {
		t.Fatalf("maybePromptedUpdate() = %v, want nil", err)
	}
	if !installCalled {
		t.Fatal("Install was not called after accepting update prompt")
	}
}
