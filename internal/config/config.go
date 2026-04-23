package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const DefaultModel = "pi/auto"

type Provider string

const (
	ProviderClaudeCode Provider = "claude-code"
	ProviderPi         Provider = "pi"
	ProviderOpenCode   Provider = "opencode"
	ProviderUnknown    Provider = "unknown"
)

func DetectProvider(model string) Provider {
	if strings.HasPrefix(model, "claude-code/") {
		return ProviderClaudeCode
	}
	if strings.HasPrefix(model, "pi/") {
		return ProviderPi
	}
	if strings.HasPrefix(model, "opencode/") || strings.HasPrefix(model, "opencode-go/") {
		return ProviderOpenCode
	}
	if strings.HasPrefix(model, "anthropic/") || strings.HasPrefix(model, "ollama/") {
		return ProviderOpenCode
	}
	return ProviderUnknown
}
const DefaultTestCommand = "go test ./..."

type Config struct {
	Model       string `json:"model"`
	PRDFile     string `json:"prd_file"`
	WorkDir     string `json:"-"`
	TestCommand string `json:"test_command"`
}

func DefaultConfig() *Config {
	return &Config{
		Model:       DefaultModel,
		PRDFile:     "prd.json",
		TestCommand: DefaultTestCommand,
	}
}

func Load() (*Config, error) {
	cfg := DefaultConfig()

	if wd, err := os.Getwd(); err == nil {
		cfg.WorkDir = wd
	}

	if err := applyEnvOverrides(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func (c *Config) ConfigPath(filename string) string {
	if c.WorkDir == "" {
		return filename
	}
	return filepath.Join(c.WorkDir, filename)
}

func (c *Config) PRDPath() string {
	return c.ConfigPath(c.PRDFile)
}

func (c *Config) ValidateModel() error {
	if c.Model == "" {
		return fmt.Errorf("model cannot be empty")
	}
	provider := DetectProvider(c.Model)
	if provider == ProviderUnknown {
		return fmt.Errorf("unknown provider for model %q (supported prefixes: claude-code/, pi/, opencode/, opencode-go/, anthropic/, ollama/)", c.Model)
	}
	if provider == ProviderPi && strings.TrimPrefix(c.Model, "pi/") == "" {
		return fmt.Errorf("model cannot be empty after pi/ prefix")
	}
	return nil
}

func (c *Config) Validate() error {
	if err := c.ValidateModel(); err != nil {
		return fmt.Errorf("invalid model configuration: %w", err)
	}
	if c.PRDFile == "" {
		return fmt.Errorf("prd_file cannot be empty")
	}
	if c.TestCommand == "" {
		return fmt.Errorf("test_command cannot be empty")
	}

	// Validate PRD file path for security (prevent path traversal)
	if filepath.Base(c.PRDFile) != c.PRDFile {
		return fmt.Errorf("prd_file must be a simple filename, got path %q", c.PRDFile)
	}
	if filepath.IsAbs(c.PRDFile) {
		return fmt.Errorf("prd_file cannot be an absolute path, got %q", c.PRDFile)
	}
	if strings.Contains(c.PRDFile, "..") {
		return fmt.Errorf("prd_file cannot contain path traversal, got %q", c.PRDFile)
	}

	return nil
}
