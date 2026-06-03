package app

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
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

func TestRunClean(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	prdPath := filepath.Join(tmpDir, "prd.json")
	if err := os.WriteFile(prdPath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("removes prd.json", func(t *testing.T) {
		if code := Run([]string{"clean"}); code != 0 {
			t.Fatalf("Run(clean) = %d, want 0", code)
		}
		if _, err := os.Stat(prdPath); !os.IsNotExist(err) {
			t.Fatalf("prd.json still exists after clean: %v", err)
		}
	})

	t.Run("skips ValidateResume with --resume", func(t *testing.T) {
		if code := Run([]string{"clean", "--resume"}); code != 0 {
			t.Fatalf("Run(clean --resume) = %d, want 0 (ValidateResume must not run)", code)
		}
	})
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
