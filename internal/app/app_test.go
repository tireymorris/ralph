package app

import (
	"os"
	"testing"
)

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
