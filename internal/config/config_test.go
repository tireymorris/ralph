package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if DefaultModel != "claude-code/sonnet" {
		t.Errorf("DefaultModel = %q, want %q", DefaultModel, "claude-code/sonnet")
	}
	if cfg.Model != DefaultModel {
		t.Errorf("Model = %q, want %q", cfg.Model, DefaultModel)
	}
	if cfg.PRDFile != "prd.json" {
		t.Errorf("PRDFile = %q, want %q", cfg.PRDFile, "prd.json")
	}
}

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		name   string
		model  string
		expect Provider
	}{
		{"claude-code prefix", "claude-code/sonnet", ProviderClaudeCode},
		{"claude-code haiku", "claude-code/haiku", ProviderClaudeCode},
		{"claude-code opus", "claude-code/opus", ProviderClaudeCode},
		{"opencode prefix", "opencode/kimi-k2.5-free", ProviderOpenCode},
		{"pi prefix", "pi/auto", ProviderPi},
		{"pi openai style", "pi/openai/gpt-4o", ProviderPi},
		{"opencode-go prefix", "opencode-go/qwen3.6-plus", ProviderOpenCode},
		{"anthropic prefix", "anthropic/claude-3-5-sonnet-20240620", ProviderOpenCode},
		{"ollama prefix", "ollama/llama3.2:3b", ProviderOpenCode},
		{"cursor-agent with model", "cursor-agent/sonnet-4", ProviderCursorAgent},
		{"cursor-agent empty suffix", "cursor-agent/", ProviderCursorAgent},
		{"unknown provider", "invalid-model", ProviderUnknown},
		{"empty string", "", ProviderUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectProvider(tt.model)
			if got != tt.expect {
				t.Errorf("DetectProvider(%q) = %v, want %v", tt.model, got, tt.expect)
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

	if cfg.Model != DefaultModel {
		t.Errorf("Model = %q, want %q", cfg.Model, DefaultModel)
	}
	if cfg.PRDFile != "prd.json" {
		t.Errorf("PRDFile = %q, want %q", cfg.PRDFile, "prd.json")
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
}

func TestLoadFullConfig(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()
	os.Setenv("RALPH_MODEL", "opencode/big-pickle")
	os.Setenv("RALPH_PRD_FILE", "custom.json")
	defer func() {
		os.Unsetenv("RALPH_MODEL")
		os.Unsetenv("RALPH_PRD_FILE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.Model != "opencode/big-pickle" {
		t.Errorf("Model = %q, want %q", cfg.Model, "opencode/big-pickle")
	}
	if cfg.PRDFile != "custom.json" {
		t.Errorf("PRDFile = %q, want %q", cfg.PRDFile, "custom.json")
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
			name:    "valid opencode model",
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
			name:    "valid anthropic model",
			model:   "anthropic/claude-3-5-sonnet-20240620",
			wantErr: false,
		},
		{
			name:    "valid ollama model",
			model:   "ollama/llama3.2:3b",
			wantErr: false,
		},
		{
			name:    "valid pi model",
			model:   "pi/auto",
			wantErr: false,
		},
		{
			name:    "valid cursor-agent model",
			model:   "cursor-agent/sonnet-4",
			wantErr: false,
		},
		{
			name:    "valid cursor-agent empty suffix",
			model:   "cursor-agent/",
			wantErr: false,
		},
		{
			name:    "invalid pi empty pattern",
			model:   "pi/",
			wantErr: true,
		},
		{
			name:    "invalid model - unknown provider",
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
			name: "invalid model - unknown provider",
			config: &Config{
				Model:       "invalid-model",
				PRDFile:     "prd.json",
				TestCommand: DefaultTestCommand,
			},
			wantErr: true,
		},
		{
			name: "empty prd_file",
			config: &Config{
				Model:       DefaultModel,
				PRDFile:     "",
				TestCommand: DefaultTestCommand,
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
}

func TestDefaultConfigTestCommand(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.TestCommand != "go test ./..." {
		t.Errorf("TestCommand = %q, want %q", cfg.TestCommand, "go test ./...")
	}
}

func TestLoadDefaultTestCommand(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.TestCommand != "go test ./..." {
		t.Errorf("TestCommand = %q, want %q", cfg.TestCommand, "go test ./...")
	}
}

func TestLoadCustomTestCommand(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()
	os.Setenv("RALPH_TEST_COMMAND", "go test -v ./internal/...")
	defer os.Unsetenv("RALPH_TEST_COMMAND")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.TestCommand != "go test -v ./internal/..." {
		t.Errorf("TestCommand = %q, want %q", cfg.TestCommand, "go test -v ./internal/...")
	}
}
