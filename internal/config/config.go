package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var SupportedModels = []string{
	"opencode/big-pickle",
	"opencode/glm-4.7-free",
	"opencode/gpt-5-nano",
	"opencode/grok-code",
	"opencode/minimax-m2.1-free",
}

const DefaultModel = "opencode/grok-code"

type Config struct {
	Model         string `json:"model"`
	MaxIterations int    `json:"max_iterations"`
	RetryAttempts int    `json:"retry_attempts"`
	RetryDelay    int    `json:"retry_delay"`
	LogLevel      string `json:"log_level"`
	PRDFile       string `json:"prd_file"`
	WorkDir       string `json:"-"` // Working directory where ralph was invoked
}

func DefaultConfig() *Config {
	return &Config{
		Model:         DefaultModel,
		MaxIterations: 50,
		RetryAttempts: 3,
		RetryDelay:    5,
		LogLevel:      "info",
		PRDFile:       "prd.json",
	}
}

func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Capture the working directory where ralph was invoked
	if wd, err := os.Getwd(); err == nil {
		cfg.WorkDir = wd
	}

	data, err := os.ReadFile(cfg.ConfigPath("ralph.config.json"))
	if err != nil {
		// If config file doesn't exist, use defaults and validate
		return cfg, cfg.Validate()
	}

	var fileCfg Config
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge file config with defaults
	if fileCfg.Model != "" {
		cfg.Model = fileCfg.Model
	}
	if fileCfg.MaxIterations > 0 {
		cfg.MaxIterations = fileCfg.MaxIterations
	}
	if fileCfg.RetryAttempts > 0 {
		cfg.RetryAttempts = fileCfg.RetryAttempts
	}
	if fileCfg.RetryDelay > 0 {
		cfg.RetryDelay = fileCfg.RetryDelay
	}
	if fileCfg.LogLevel != "" {
		cfg.LogLevel = fileCfg.LogLevel
	}
	if fileCfg.PRDFile != "" {
		cfg.PRDFile = fileCfg.PRDFile
	}

	// Validate the merged config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// ConfigPath returns the full path to a file in the working directory
func (c *Config) ConfigPath(filename string) string {
	if c.WorkDir == "" {
		return filename
	}
	return filepath.Join(c.WorkDir, filename)
}

// PRDPath returns the full path to the PRD file
func (c *Config) PRDPath() string {
	return c.ConfigPath(c.PRDFile)
}

// ValidateModel checks if the configured model is in the list of supported models.
// Returns an error if the model is not supported.
func (c *Config) ValidateModel() error {
	for _, m := range SupportedModels {
		if c.Model == m {
			return nil
		}
	}
	return fmt.Errorf("unsupported model: %s (supported: %v)", c.Model, SupportedModels)
}

// Validate checks all configuration values for validity
func (c *Config) Validate() error {
	if err := c.ValidateModel(); err != nil {
		return err
	}
	if c.MaxIterations <= 0 {
		return fmt.Errorf("max_iterations must be positive, got %d", c.MaxIterations)
	}
	if c.RetryAttempts < 0 {
		return fmt.Errorf("retry_attempts must be non-negative, got %d", c.RetryAttempts)
	}
	if c.RetryDelay < 0 {
		return fmt.Errorf("retry_delay must be non-negative, got %d", c.RetryDelay)
	}
	if c.LogLevel == "" {
		return fmt.Errorf("log_level cannot be empty")
	}
	if c.PRDFile == "" {
		return fmt.Errorf("prd_file cannot be empty")
	}
	return nil
}
