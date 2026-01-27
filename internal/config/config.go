package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var SupportedModels = []string{
	"opencode/big-pickle",
	"claude-code/sonnet",
	"claude-code/haiku",
	"claude-code/opus",
}

const DefaultModel = "opencode/big-pickle"

type Config struct {
	Model         string `json:"model"`
	MaxIterations int    `json:"max_iterations"`
	RetryAttempts int    `json:"retry_attempts"`
	PRDFile       string `json:"prd_file"`
	WorkDir       string `json:"-"`
}

func DefaultConfig() *Config {
	return &Config{
		Model:         DefaultModel,
		MaxIterations: 50,
		RetryAttempts: 3,
		PRDFile:       "prd.json",
	}
}

func Load() (*Config, error) {
	cfg := DefaultConfig()

	if wd, err := os.Getwd(); err == nil {
		cfg.WorkDir = wd
	}

	configPath := cfg.ConfigPath("ralph.config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		// If config file doesn't exist, return default config
		if os.IsNotExist(err) {
			return cfg, cfg.Validate()
		}
		return nil, fmt.Errorf("failed to read config file %q: %w", configPath, err)
	}

	var fileCfg Config
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %q: %w", configPath, err)
	}

	if fileCfg.Model != "" {
		cfg.Model = fileCfg.Model
	}
	if fileCfg.MaxIterations > 0 {
		cfg.MaxIterations = fileCfg.MaxIterations
	}
	if fileCfg.RetryAttempts > 0 {
		cfg.RetryAttempts = fileCfg.RetryAttempts
	}
	if fileCfg.PRDFile != "" {
		cfg.PRDFile = fileCfg.PRDFile
	}

	if err := cfg.Validate(); err != nil {
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
	for _, m := range SupportedModels {
		if c.Model == m {
			return nil
		}
	}
	return fmt.Errorf("unsupported model: %s (supported: %v)", c.Model, SupportedModels)
}

func (c *Config) Validate() error {
	if err := c.ValidateModel(); err != nil {
		return fmt.Errorf("invalid model configuration: %w", err)
	}
	if c.MaxIterations <= 0 {
		return fmt.Errorf("max_iterations must be positive, got %d", c.MaxIterations)
	}
	if c.RetryAttempts < 0 {
		return fmt.Errorf("retry_attempts must be non-negative, got %d", c.RetryAttempts)
	}
	if c.PRDFile == "" {
		return fmt.Errorf("prd_file cannot be empty")
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
