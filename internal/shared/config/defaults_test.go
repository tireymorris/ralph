package config

import (
	"os"
	"testing"
)

func TestDefaultConfigBranchPrefix(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.BranchPrefix != DefaultBranchPrefix {
		t.Fatalf("BranchPrefix = %q, want %q", cfg.BranchPrefix, DefaultBranchPrefix)
	}
}

func TestDefaultConfigTestCommandEmpty(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.TestCommand != "" {
		t.Fatalf("TestCommand = %q, want empty default", cfg.TestCommand)
	}
}

func TestLoadEnvBranchPrefix(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()
	os.Setenv("RALPH_BRANCH_PREFIX", "feat")
	defer os.Unsetenv("RALPH_BRANCH_PREFIX")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.BranchPrefix != "feat" {
		t.Fatalf("BranchPrefix = %q, want feat", cfg.BranchPrefix)
	}
}

func TestLoadEnvDefaultBranches(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()
	os.Setenv("RALPH_DEFAULT_BRANCHES", "trunk,develop")
	defer os.Unsetenv("RALPH_DEFAULT_BRANCHES")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.DefaultBranches) != 2 || cfg.DefaultBranches[0] != "trunk" || cfg.DefaultBranches[1] != "develop" {
		t.Fatalf("DefaultBranches = %v, want [trunk develop]", cfg.DefaultBranches)
	}
}

func TestLoadEnvTestCommand(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()
	os.Setenv("RALPH_TEST_COMMAND", "npm test")
	defer os.Unsetenv("RALPH_TEST_COMMAND")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.TestCommand != "npm test" {
		t.Fatalf("TestCommand = %q, want npm test", cfg.TestCommand)
	}
}

func TestValidateAllowsEmptyTestCommand(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TestCommand = ""
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil for empty test command", err)
	}
}

func TestLoadDetectsTestCommandFromWorkdir(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	if err := os.WriteFile(tmpDir+"/go.mod", []byte("module example.com/app\n"), 0644); err != nil {
		t.Fatal(err)
	}
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.TestCommand != "go test ./..." {
		t.Fatalf("TestCommand = %q, want go test ./...", cfg.TestCommand)
	}
}
