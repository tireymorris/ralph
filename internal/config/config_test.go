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

	// Test Claude Code models are present
	claudeModels := []string{
		"claude-code/sonnet",
		"claude-code/haiku",
		"claude-code/opus",
	}

	for _, model := range claudeModels {
		found := false
		for _, supported := range SupportedModels {
			if supported == model {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Claude Code model %q not in SupportedModels", model)
		}
	}

	// Test OpenCode default model is present
	found = false
	for _, supported := range SupportedModels {
		if supported == "opencode/big-pickle" {
			found = true
			break
		}
	}
	if !found {
		t.Error("opencode/big-pickle not in SupportedModels")
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

	if cfg.Model != DefaultModel {
		t.Errorf("Model = %q, want %q", cfg.Model, DefaultModel)
	}
	if cfg.MaxIterations != 50 {
		t.Errorf("MaxIterations = %d, want 50", cfg.MaxIterations)
	}
	if cfg.RetryAttempts != 3 {
		t.Errorf("RetryAttempts = %d, want 3", cfg.RetryAttempts)
	}
	if cfg.PRDFile != "prd.json" {
		t.Errorf("PRDFile = %q, want %q", cfg.PRDFile, "prd.json")
	}
}

func TestLoadInvalidMaxIterations(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()
	os.Setenv("RALPH_MAX_ITERATIONS", "invalid")
	defer os.Unsetenv("RALPH_MAX_ITERATIONS")

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for invalid RALPH_MAX_ITERATIONS")
	}
}

func TestLoadPartialConfig(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()
	os.Setenv("RALPH_MODEL", "opencode/big-pickle")
	defer func() {
		os.Unsetenv("RALPH_MODEL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.Model != "opencode/big-pickle" {
		t.Errorf("Model = %q, want %q", cfg.Model, "opencode/big-pickle")
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

	os.Clearenv()
	os.Setenv("RALPH_MODEL", "opencode/big-pickle")
	os.Setenv("RALPH_MAX_ITERATIONS", "100")
	os.Setenv("RALPH_RETRY_ATTEMPTS", "5")
	os.Setenv("RALPH_PRD_FILE", "custom.json")
	defer func() {
		os.Unsetenv("RALPH_MODEL")
		os.Unsetenv("RALPH_MAX_ITERATIONS")
		os.Unsetenv("RALPH_RETRY_ATTEMPTS")
		os.Unsetenv("RALPH_PRD_FILE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.Model != "opencode/big-pickle" {
		t.Errorf("Model = %q, want %q", cfg.Model, "opencode/big-pickle")
	}
	if cfg.MaxIterations != 100 {
		t.Errorf("MaxIterations = %d, want 100", cfg.MaxIterations)
	}
	if cfg.RetryAttempts != 5 {
		t.Errorf("RetryAttempts = %d, want 5", cfg.RetryAttempts)
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

	os.Clearenv()
	os.Setenv("RALPH_MAX_ITERATIONS", "0")
	defer os.Unsetenv("RALPH_MAX_ITERATIONS")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.MaxIterations != 50 {
		t.Errorf("MaxIterations = %d, want default 50", cfg.MaxIterations)
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

	// Resolve symlinks for comparison (handles macOS /var -> /private/var)
	wantDir, _ := filepath.EvalSymlinks(tmpDir)
	gotDir, _ := filepath.EvalSymlinks(cfg.WorkDir)

	if gotDir != wantDir {
		t.Errorf("WorkDir = %q, want %q", cfg.WorkDir, tmpDir)
	}
}

func TestConfigPath(t *testing.T) {
	cfg := &Config{
		WorkDir: "/some/path",
		PRDFile: "prd.json",
	}

	got := cfg.ConfigPath("test.json")
	want := filepath.Join("/some/path", "test.json")
	if got != want {
		t.Errorf("ConfigPath() = %q, want %q", got, want)
	}
}

func TestConfigPathEmptyWorkDir(t *testing.T) {
	cfg := &Config{
		WorkDir: "",
		PRDFile: "prd.json",
	}

	got := cfg.ConfigPath("test.json")
	want := "test.json"
	if got != want {
		t.Errorf("ConfigPath() = %q, want %q", got, want)
	}
}

func TestPRDPath(t *testing.T) {
	cfg := &Config{
		WorkDir: "/some/path",
		PRDFile: "custom.json",
	}

	got := cfg.PRDPath()
	want := filepath.Join("/some/path", "custom.json")
	if got != want {
		t.Errorf("PRDPath() = %q, want %q", got, want)
	}
}

func TestValidateModel(t *testing.T) {
	tests := []struct {
		name    string
		model   string
		wantErr bool
	}{
		{
			name:    "valid default model",
			model:   DefaultModel,
			wantErr: false,
		},
		{
			name:    "valid supported model",
			model:   "opencode/big-pickle",
			wantErr: false,
		},
		{
			name:    "valid claude code model - sonnet",
			model:   "claude-code/sonnet",
			wantErr: false,
		},
		{
			name:    "valid claude code model - haiku",
			model:   "claude-code/haiku",
			wantErr: false,
		},
		{
			name:    "valid claude code model - opus",
			model:   "claude-code/opus",
			wantErr: false,
		},
		{
			name:    "invalid model",
			model:   "invalid-model",
			wantErr: true,
		},
		{
			name:    "empty model",
			model:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Model: tt.model}
			err := cfg.ValidateModel()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateModel() error = %v, wantErr %v", err, tt.wantErr)
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
		{
			name:    "valid default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid model",
			config: &Config{
				Model:         "invalid-model",
				MaxIterations: 50,
				RetryAttempts: 3,
				PRDFile:       "prd.json",
			},
			wantErr: true,
		},
		{
			name: "negative max_iterations",
			config: &Config{
				Model:         DefaultModel,
				MaxIterations: -1,
				RetryAttempts: 3,
				PRDFile:       "prd.json",
			},
			wantErr: true,
		},
		{
			name: "zero max_iterations",
			config: &Config{
				Model:         DefaultModel,
				MaxIterations: 0,
				RetryAttempts: 3,
				PRDFile:       "prd.json",
			},
			wantErr: true,
		},
		{
			name: "negative retry_attempts",
			config: &Config{
				Model:         DefaultModel,
				MaxIterations: 50,
				RetryAttempts: -1,
				PRDFile:       "prd.json",
			},
			wantErr: true,
		},
		{
			name: "empty prd_file",
			config: &Config{
				Model:         DefaultModel,
				MaxIterations: 50,
				RetryAttempts: 3,
				PRDFile:       "",
			},
			wantErr: true,
		},
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
	os.Setenv("RALPH_MODEL", "invalid-model")
	defer os.Unsetenv("RALPH_MODEL")

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for invalid model")
	}
}

func TestLoadClaudeCodeConfig(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	tests := []struct {
		name    string
		model   string
		wantErr bool
	}{
		{
			name:    "claude sonnet model",
			model:   "claude-code/sonnet",
			wantErr: false,
		},
		{
			name:    "claude haiku model",
			model:   "claude-code/haiku",
			wantErr: false,
		},
		{
			name:    "claude opus model",
			model:   "claude-code/opus",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv("RALPH_MODEL", tt.model)
			defer os.Unsetenv("RALPH_MODEL")

			cfg, err := Load()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if cfg.Model != tt.model {
					t.Errorf("Model = %q, want %q", cfg.Model, tt.model)
				}
				if cfg.MaxIterations != 50 {
					t.Errorf("MaxIterations = %d, want default 50", cfg.MaxIterations)
				}
				if cfg.RetryAttempts != 3 {
					t.Errorf("RetryAttempts = %d, want default 3", cfg.RetryAttempts)
				}
			}
		})
	}
}

func TestLoadCursorAgentConfig(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()
	os.Setenv("RALPH_MODEL", "claude-code/sonnet")
	defer os.Unsetenv("RALPH_MODEL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.Model != "claude-code/sonnet" {
		t.Errorf("Model = %q, want %q", cfg.Model, "claude-code/sonnet")
	}
	if cfg.MaxIterations != 50 {
		t.Errorf("MaxIterations = %d, want default 50", cfg.MaxIterations)
	}
	if cfg.RetryAttempts != 3 {
		t.Errorf("RetryAttempts = %d, want default 3", cfg.RetryAttempts)
	}
}

func TestLoadInvalidRetryAttempts(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()
	os.Setenv("RALPH_RETRY_ATTEMPTS", "invalid")
	defer os.Unsetenv("RALPH_RETRY_ATTEMPTS")

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for invalid RALPH_RETRY_ATTEMPTS")
	}
}
