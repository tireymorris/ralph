package main

import (
	"os"
	"testing"

	"ralph/internal/app"
	"ralph/internal/args"
	"ralph/internal/shared/config"
)

func TestRunHelp(t *testing.T) {
	if code := app.Run([]string{"--help"}); code != 0 {
		t.Errorf("app.Run() with --help = %d, want 0", code)
	}
}

func TestRunStatus(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	defer os.Chdir(origDir)

	if code := app.Run([]string{"status"}); code != 0 {
		t.Errorf("app.Run() with status = %d, want 0", code)
	}
}

func TestValidateResume(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{PRDFile: "prd.json", WorkDir: tmpDir}
	if err := app.ValidateResume(cfg, false); err != nil {
		t.Fatalf("ValidateResume(false) = %v", err)
	}
	if err := app.ValidateResume(cfg, true); err == nil {
		t.Fatal("ValidateResume(true) want error")
	}
}

func TestArgsStatus(t *testing.T) {
	opts := args.Parse([]string{"status"})
	if !opts.Status {
		t.Fatal("status not parsed")
	}
}
