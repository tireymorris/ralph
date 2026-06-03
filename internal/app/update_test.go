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
	defer func() { updateInstall = oldInstall }()
	updateInstall = func(ctx context.Context, opts update.InstallOptions) error {
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
	if !strings.Contains(buf.String(), "bin/ralph") {
		t.Errorf("stdout = %q, want substring bin/ralph", buf.String())
	}
}

func TestRunUpdateFailure(t *testing.T) {
	oldInstall := updateInstall
	defer func() { updateInstall = oldInstall }()
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
