package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunHelp(t *testing.T) {
	origArgs := os.Args
	os.Args = []string{"ralph", "--help"}
	defer func() { os.Args = origArgs }()

	code := run()
	if code != 0 {
		t.Errorf("run() with --help = %d, want 0", code)
	}
}

func TestRunHelpShort(t *testing.T) {
	origArgs := os.Args
	os.Args = []string{"ralph", "-h"}
	defer func() { os.Args = origArgs }()

	code := run()
	if code != 0 {
		t.Errorf("run() with -h = %d, want 0", code)
	}
}

func TestRunNoArgs(t *testing.T) {
	origArgs := os.Args
	os.Args = []string{"ralph"}
	defer func() { os.Args = origArgs }()

	code := run()
	if code != 1 {
		t.Errorf("run() with no args = %d, want 1", code)
	}
}

func TestRunResumeNoPRD(t *testing.T) {
	origArgs := os.Args
	origDir, _ := os.Getwd()

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	os.Args = []string{"ralph", "--resume"}
	defer func() {
		os.Args = origArgs
		os.Chdir(origDir)
	}()

	code := run()
	if code != 1 {
		t.Errorf("run() with --resume and no prd = %d, want 1", code)
	}
}

func TestRunResumeInvalidPRD(t *testing.T) {
	origArgs := os.Args
	origDir, _ := os.Getwd()

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	os.WriteFile("prd.json", []byte("invalid json"), 0644)

	os.Args = []string{"ralph", "--resume"}
	defer func() {
		os.Args = origArgs
		os.Chdir(origDir)
	}()

	code := run()
	if code != 1 {
		t.Errorf("run() with invalid prd = %d, want 1", code)
	}
}

func TestRunResumeValidPRDHeadless(t *testing.T) {
	origArgs := os.Args
	origDir, _ := os.Getwd()

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	prdContent := `{"project_name":"Test","stories":[{"id":"1","title":"T","description":"D","acceptance_criteria":["a"],"priority":1,"passes":true}]}`
	os.WriteFile("prd.json", []byte(prdContent), 0644)

	os.Args = []string{"ralph", "run", "--resume"}
	defer func() {
		os.Args = origArgs
		os.Chdir(origDir)
	}()

	code := run()
	if code != 0 {
		t.Errorf("run() with valid prd (all complete) = %d, want 0", code)
	}
}

func TestRunDryRunHeadless(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that requires opencode")
	}
	origArgs := os.Args
	origDir, _ := os.Getwd()

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	os.WriteFile(filepath.Join(tmpDir, "ralph.config.json"), []byte(`{"model":"test"}`), 0644)

	os.Args = []string{"ralph", "run", "test", "--dry-run"}
	defer func() {
		os.Args = origArgs
		os.Chdir(origDir)
	}()
}
