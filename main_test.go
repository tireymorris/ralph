package main

import (
	"os"
	"testing"

	"ralph/internal/args"
	"ralph/internal/shared/config"
)

func TestRunHelp(t *testing.T) {
	origArgs := os.Args
	origLaunch := launchTUI
	launchTUI = func(*config.Config, *args.Options) int { return 0 }
	os.Args = []string{"ralph", "--help"}
	defer func() { os.Args = origArgs; launchTUI = origLaunch }()

	if code := run(); code != 0 {
		t.Errorf("run() with --help = %d, want 0", code)
	}
}

func TestRunHelpShort(t *testing.T) {
	origArgs := os.Args
	origLaunch := launchTUI
	launchTUI = func(*config.Config, *args.Options) int { return 0 }
	os.Args = []string{"ralph", "-h"}
	defer func() { os.Args = origArgs; launchTUI = origLaunch }()

	if code := run(); code != 0 {
		t.Errorf("run() with -h = %d, want 0", code)
	}
}

func TestRunNoArgs(t *testing.T) {
	origArgs := os.Args
	origLaunch := launchTUI
	launchTUI = func(*config.Config, *args.Options) int { return 1 }
	os.Args = []string{"ralph"}
	defer func() { os.Args = origArgs; launchTUI = origLaunch }()

	if code := run(); code != 1 {
		t.Errorf("run() with no args = %d, want 1", code)
	}
}

func TestRunResumeNoPRD(t *testing.T) {
	origArgs := os.Args
	origDir, _ := os.Getwd()
	origModel := os.Getenv("RALPH_MODEL")
	origLaunch := launchTUI
	launchTUI = func(*config.Config, *args.Options) int { return 0 }

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	os.Setenv("RALPH_MODEL", "opencode/big-pickle")
	os.Args = []string{"ralph", "--resume"}
	defer func() {
		os.Args = origArgs
		os.Chdir(origDir)
		launchTUI = origLaunch
		if origModel != "" {
			os.Setenv("RALPH_MODEL", origModel)
		} else {
			os.Unsetenv("RALPH_MODEL")
		}
	}()

	if code := run(); code != 1 {
		t.Errorf("run() with --resume and no prd = %d, want 1", code)
	}
}

func TestRunResumeInvalidPRD(t *testing.T) {
	origArgs := os.Args
	origDir, _ := os.Getwd()
	origModel := os.Getenv("RALPH_MODEL")
	origLaunch := launchTUI
	launchTUI = func(*config.Config, *args.Options) int { return 0 }

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	os.Setenv("RALPH_MODEL", "opencode/big-pickle")
	os.WriteFile("prd.json", []byte("invalid json"), 0644)
	os.Args = []string{"ralph", "--resume"}
	defer func() {
		os.Args = origArgs
		os.Chdir(origDir)
		launchTUI = origLaunch
		if origModel != "" {
			os.Setenv("RALPH_MODEL", origModel)
		} else {
			os.Unsetenv("RALPH_MODEL")
		}
	}()

	if code := run(); code != 1 {
		t.Errorf("run() with invalid prd = %d, want 1", code)
	}
}

func TestRunResumeValidPRD(t *testing.T) {
	origArgs := os.Args
	origDir, _ := os.Getwd()
	origModel := os.Getenv("RALPH_MODEL")
	origLaunch := launchTUI
	launchTUI = func(*config.Config, *args.Options) int { return 0 }

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	os.Setenv("RALPH_MODEL", "opencode/big-pickle")
	os.WriteFile("prd.json", []byte(`{"project_name":"Test","stories":[{"id":"1","title":"T","description":"D","acceptance_criteria":["a"],"priority":1,"passes":true}]}`), 0644)
	os.Args = []string{"ralph", "--resume"}
	defer func() {
		os.Args = origArgs
		os.Chdir(origDir)
		launchTUI = origLaunch
		if origModel != "" {
			os.Setenv("RALPH_MODEL", origModel)
		} else {
			os.Unsetenv("RALPH_MODEL")
		}
	}()

	if code := run(); code != 0 {
		t.Errorf("run() with valid prd (all complete) = %d, want 0", code)
	}
}

func TestRunStatus(t *testing.T) {
	origArgs := os.Args
	origDir, _ := os.Getwd()
	origModel := os.Getenv("RALPH_MODEL")
	origLaunch := launchTUI
	launchTUI = func(*config.Config, *args.Options) int { return 99 }

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	os.Setenv("RALPH_MODEL", "opencode/big-pickle")
	os.Args = []string{"ralph", "status"}
	defer func() {
		os.Args = origArgs
		os.Chdir(origDir)
		launchTUI = origLaunch
		if origModel != "" {
			os.Setenv("RALPH_MODEL", origModel)
		} else {
			os.Unsetenv("RALPH_MODEL")
		}
	}()

	if code := run(); code != 0 {
		t.Errorf("run() with status = %d, want 0", code)
	}
}

func TestRunLaunchesTUI(t *testing.T) {
	origArgs := os.Args
	origDir, _ := os.Getwd()
	origModel := os.Getenv("RALPH_MODEL")
	origLaunch := launchTUI
	called := false
	launchTUI = func(_ *config.Config, opts *args.Options) int {
		called = true
		if opts.Resume || opts.Status {
			t.Fatal("unexpected flags")
		}
		return 0
	}

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	os.Setenv("RALPH_MODEL", "opencode/big-pickle")
	os.Args = []string{"ralph", "build a todo app"}
	defer func() {
		os.Args = origArgs
		os.Chdir(origDir)
		launchTUI = origLaunch
		if origModel != "" {
			os.Setenv("RALPH_MODEL", origModel)
		} else {
			os.Unsetenv("RALPH_MODEL")
		}
	}()

	if code := run(); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
	if !called {
		t.Fatal("launchTUI not called")
	}
}
