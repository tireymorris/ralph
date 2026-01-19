package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Model != DefaultModel {
		t.Errorf("Model = %q, want %q", cfg.Model, DefaultModel)
	}
	if cfg.MaxIterations != 50 {
		t.Errorf("MaxIterations = %d, want 50", cfg.MaxIterations)
	}
	if cfg.RetryAttempts != 3 {
		t.Errorf("RetryAttempts = %d, want 3", cfg.RetryAttempts)
	}
	if cfg.RetryDelay != 5 {
		t.Errorf("RetryDelay = %d, want 5", cfg.RetryDelay)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}
	if cfg.PRDFile != "prd.json" {
		t.Errorf("PRDFile = %q, want %q", cfg.PRDFile, "prd.json")
	}
}

func TestSupportedModels(t *testing.T) {
	if len(SupportedModels) == 0 {
		t.Error("SupportedModels should not be empty")
	}

	found := false
	for _, m := range SupportedModels {
		if m == DefaultModel {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("DefaultModel %q not in SupportedModels", DefaultModel)
	}
}

func TestLoadNoFile(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	cfg := Load()

	if cfg.Model != DefaultModel {
		t.Errorf("Model = %q, want %q", cfg.Model, DefaultModel)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.WriteFile("ralph.config.json", []byte("invalid json"), 0644)

	cfg := Load()

	if cfg.Model != DefaultModel {
		t.Errorf("Model = %q, want default %q", cfg.Model, DefaultModel)
	}
}

func TestLoadPartialConfig(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	configContent := `{"model": "custom-model"}`
	os.WriteFile("ralph.config.json", []byte(configContent), 0644)

	cfg := Load()

	if cfg.Model != "custom-model" {
		t.Errorf("Model = %q, want %q", cfg.Model, "custom-model")
	}
	if cfg.MaxIterations != 50 {
		t.Errorf("MaxIterations = %d, want default 50", cfg.MaxIterations)
	}
}

func TestLoadFullConfig(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	configContent := `{
		"model": "test-model",
		"max_iterations": 100,
		"retry_attempts": 5,
		"retry_delay": 10,
		"log_level": "debug",
		"prd_file": "custom.json"
	}`
	os.WriteFile(filepath.Join(tmpDir, "ralph.config.json"), []byte(configContent), 0644)

	cfg := Load()

	if cfg.Model != "test-model" {
		t.Errorf("Model = %q, want %q", cfg.Model, "test-model")
	}
	if cfg.MaxIterations != 100 {
		t.Errorf("MaxIterations = %d, want 100", cfg.MaxIterations)
	}
	if cfg.RetryAttempts != 5 {
		t.Errorf("RetryAttempts = %d, want 5", cfg.RetryAttempts)
	}
	if cfg.RetryDelay != 10 {
		t.Errorf("RetryDelay = %d, want 10", cfg.RetryDelay)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
	if cfg.PRDFile != "custom.json" {
		t.Errorf("PRDFile = %q, want %q", cfg.PRDFile, "custom.json")
	}
}

func TestLoadZeroValuesIgnored(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	configContent := `{
		"model": "",
		"max_iterations": 0,
		"retry_attempts": 0,
		"retry_delay": 0,
		"log_level": "",
		"prd_file": ""
	}`
	os.WriteFile("ralph.config.json", []byte(configContent), 0644)

	cfg := Load()

	if cfg.Model != DefaultModel {
		t.Errorf("Model = %q, want default %q", cfg.Model, DefaultModel)
	}
	if cfg.MaxIterations != 50 {
		t.Errorf("MaxIterations = %d, want default 50", cfg.MaxIterations)
	}
	if cfg.RetryAttempts != 3 {
		t.Errorf("RetryAttempts = %d, want default 3", cfg.RetryAttempts)
	}
	if cfg.RetryDelay != 5 {
		t.Errorf("RetryDelay = %d, want default 5", cfg.RetryDelay)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want default %q", cfg.LogLevel, "info")
	}
	if cfg.PRDFile != "prd.json" {
		t.Errorf("PRDFile = %q, want default %q", cfg.PRDFile, "prd.json")
	}
}
