package app

import (
	"bytes"
	"io"
	"os"
	"testing"

	"ralph/internal/version"
)

func TestRunVersion(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	codeCh := make(chan int, 1)
	go func() {
		codeCh <- Run([]string{"version"})
		w.Close()
	}()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	if code := <-codeCh; code != 0 {
		t.Errorf("Run(version) = %d, want 0", code)
	}
	want := version.Info() + "\n"
	if got := buf.String(); got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunBareNoTTY(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin
		r.Close()
		w.Close()
	}()
	w.Close()

	if code := Run([]string{}); code != 1 {
		t.Errorf("Run() with no args and non-TTY stdin = %d, want 1", code)
	}
}
