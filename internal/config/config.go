package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var SupportedModels = []string{
	"opencode/kimi-k2.5-free",
	"opencode/big-pickle",
	"opencode/glm-4.7-free",
	"opencode/gpt-5-nano",
	"opencode/minimax-m2.1-free",
	"opencode/trinity-large-preview-free",
	"claude-code/sonnet",
	"claude-code/haiku",
	"claude-code/opus",
}

const DefaultModel = "opencode/kimi-k2.5-free"

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

	if model := os.Getenv("RALPH_MODEL"); model != "" {
		cfg.Model = model
	}

	if maxIterStr := os.Getenv("RALPH_MAX_ITERATIONS"); maxIterStr != "" {
		maxIter, err := strconv.Atoi(maxIterStr)
		if err != nil {
			return nil, fmt.Errorf("invalid RALPH_MAX_ITERATIONS value %q: %w", maxIterStr, err)
		}
		if maxIter > 0 {
			cfg.MaxIterations = maxIter
		}
	}

	if retryStr := os.Getenv("RALPH_RETRY_ATTEMPTS"); retryStr != "" {
		retry, err := strconv.Atoi(retryStr)
		if err != nil {
			return nil, fmt.Errorf("invalid RALPH_RETRY_ATTEMPTS value %q: %w", retryStr, err)
		}
		if retry >= 0 {
			cfg.RetryAttempts = retry
		}
	}

	if prdFile := os.Getenv("RALPH_PRD_FILE"); prdFile != "" {
		cfg.PRDFile = prdFile
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
