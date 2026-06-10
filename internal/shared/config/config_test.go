package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if DefaultRunner != "claude" {
		t.Errorf("DefaultRunner = %q, want %q", DefaultRunner, "claude")
	}
	if cfg.Runner != DefaultRunner {
		t.Errorf("Runner = %q, want %q", cfg.Runner, DefaultRunner)
	}
	if cfg.PRDFile != "prd.json" {
		t.Errorf("PRDFile = %q, want %q", cfg.PRDFile, "prd.json")
	}
}

func TestDetectRunner(t *testing.T) {
	tests := []struct {
		runner string
		want   RunnerKind
	}{
		{"claude", RunnerClaude},
		{"cursor", RunnerCursor},
		{"pi", RunnerPi},
		{"opencode", RunnerOpenCode},
		{"mock", RunnerMock},
		{"invalid-runner", RunnerUnknown},
		{"", RunnerUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.runner, func(t *testing.T) {
			got := DetectRunner(tt.runner)
			if got != tt.want {
				t.Errorf("DetectRunner(%q) = %v, want %v", tt.runner, got, tt.want)
			}
		})
	}
}

func TestLoadDefaults(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.Runner != DefaultRunner {
		t.Errorf("Runner = %q, want %q", cfg.Runner, DefaultRunner)
	}
	if cfg.PRDFile != "prd.json" {
		t.Errorf("PRDFile = %q, want %q", cfg.PRDFile, "prd.json")
	}
}

func TestLoadEnvRunner(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()
	os.Setenv("RALPH_RUNNER", "opencode")
	defer os.Unsetenv("RALPH_RUNNER")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.Runner != "opencode" {
		t.Errorf("Runner = %q, want %q", cfg.Runner, "opencode")
	}
}

func TestLoadEnvRunnerTimeout(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()
	os.Setenv("RALPH_RUNNER_TIMEOUT", "30m")
	defer os.Unsetenv("RALPH_RUNNER_TIMEOUT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.RunnerTimeout != 30*time.Minute {
		t.Errorf("RunnerTimeout = %v, want %v", cfg.RunnerTimeout, 30*time.Minute)
	}
}

func TestLoadEnvRunnerTimeoutRejectsInvalidDuration(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()
	os.Setenv("RALPH_RUNNER_TIMEOUT", "eventually")
	defer os.Unsetenv("RALPH_RUNNER_TIMEOUT")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "RALPH_RUNNER_TIMEOUT") {
		t.Errorf("Load() error = %v, want mention RALPH_RUNNER_TIMEOUT", err)
	}
}

func TestLoadEnvYoloEnablesAutoApprove(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()
	os.Setenv("RALPH_YOLO", "1")
	defer os.Unsetenv("RALPH_YOLO")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if !cfg.AutoApprove {
		t.Error("AutoApprove should be true when RALPH_YOLO=1")
	}
}

func TestLoadSetsWorkDir(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	wantDir, _ := filepath.EvalSymlinks(tmpDir)
	gotDir, _ := filepath.EvalSymlinks(cfg.WorkDir)

	if gotDir != wantDir {
		t.Errorf("WorkDir = %q, want %q", cfg.WorkDir, tmpDir)
	}
}

func TestConfigPath(t *testing.T) {
	cfg := &Config{WorkDir: "/some/path", PRDFile: "prd.json"}

	got := cfg.ConfigPath("test.json")
	want := filepath.Join("/some/path", "test.json")
	if got != want {
		t.Errorf("ConfigPath() = %q, want %q", got, want)
	}
}

func TestConfigPathEmptyWorkDir(t *testing.T) {
	cfg := &Config{WorkDir: "", PRDFile: "prd.json"}

	got := cfg.ConfigPath("test.json")
	want := "test.json"
	if got != want {
		t.Errorf("ConfigPath() = %q, want %q", got, want)
	}
}

func TestPRDPath(t *testing.T) {
	cfg := &Config{WorkDir: "/some/path", PRDFile: "custom.json"}

	got := cfg.PRDPath()
	want := filepath.Join("/some/path", "custom.json")
	if got != want {
		t.Errorf("PRDPath() = %q, want %q", got, want)
	}
}

func TestValidateRunner(t *testing.T) {
	tests := []struct {
		runner  string
		wantErr bool
	}{
		{DefaultRunner, false},
		{"opencode", false},
		{"claude", false},
		{"pi", false},
		{"cursor", false},
		{"pi/", true},
		{"invalid-runner", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.runner, func(t *testing.T) {
			cfg := &Config{Runner: tt.runner}
			err := cfg.ValidateRunner()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRunner() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{name: "valid default config", config: DefaultConfig()},
		{name: "invalid runner", config: &Config{Runner: "invalid-runner", PRDFile: "prd.json", TestCommand: DefaultTestCommand}, wantErr: true},
		{name: "empty prd_file", config: &Config{Runner: DefaultRunner, PRDFile: "", TestCommand: DefaultTestCommand}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadInvalidConfig(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()
	os.Setenv("RALPH_RUNNER", "invalid-runner")
	defer os.Unsetenv("RALPH_RUNNER")

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for invalid runner")
	}
}

func TestDefaultConfigTestCommand(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.TestCommand != "go test ./..." {
		t.Errorf("TestCommand = %q, want %q", cfg.TestCommand, "go test ./...")
	}
}

func TestDefaultConfigSkipCleanupIsFalse(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.SkipCleanup {
		t.Error("SkipCleanup should default to false")
	}
}
