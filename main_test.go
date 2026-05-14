package main

import (
	"os"
	"testing"

	"ralph/internal/args"
	"ralph/internal/shared/config"
)

func TestRunHelp(t *testing.T) {
	origArgs := os.Args
	os.Args = []string{"ralph", "--help"}
	defer func() { os.Args = origArgs }()
	if code := run(); code != 0 {
		t.Errorf("run() with --help = %d, want 0", code)
	}
}

func TestRunStatus(t *testing.T) {
	origArgs := os.Args
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	os.Args = []string{"ralph", "status"}
	defer func() { os.Args = origArgs; os.Chdir(origDir) }()
	if code := run(); code != 0 {
		t.Errorf("run() with status = %d, want 0", code)
	}
}

func TestValidateResume(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{PRDFile: "prd.json", WorkDir: tmpDir}
	if err := validateResume(cfg, false); err != nil {
		t.Fatalf("ValidateResume(false) = %v", err)
	}
	if err := validateResume(cfg, true); err == nil {
		t.Fatal("ValidateResume(true) want error")
	}
}

func TestArgsStatus(t *testing.T) {
	opts := args.Parse([]string{"status"})
	if !opts.Status {
		t.Fatal("status not parsed")
	}
}
