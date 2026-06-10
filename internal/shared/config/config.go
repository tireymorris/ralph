package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const DefaultRunner = "claude"

type RunnerKind string

const (
	RunnerClaude   RunnerKind = "claude"
	RunnerCursor   RunnerKind = "cursor"
	RunnerPi       RunnerKind = "pi"
	RunnerOpenCode RunnerKind = "opencode"
	RunnerMock     RunnerKind = "mock"
	RunnerUnknown  RunnerKind = "unknown"
)

func DetectRunner(runner string) RunnerKind {
	switch runner {
	case string(RunnerClaude):
		return RunnerClaude
	case string(RunnerCursor):
		return RunnerCursor
	case string(RunnerPi):
		return RunnerPi
	case string(RunnerOpenCode):
		return RunnerOpenCode
	case string(RunnerMock):
		return RunnerMock
	default:
		return RunnerUnknown
	}
}

const DefaultTestCommand = "go test ./..."

type Config struct {
	Runner      string `json:"runner"`
	PRDFile     string `json:"prd_file"`
	WorkDir     string `json:"-"`
	TestCommand   string        `json:"test_command"`
	RunnerTimeout time.Duration `json:"-"`
	SkipCleanup   bool          `json:"-"`
	AutoApprove   bool          `json:"-"`
}

func DefaultConfig() *Config {
	return &Config{
		Runner:      DefaultRunner,
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

func (c *Config) ValidateRunner() error {
	if c.Runner == "" {
		return errors.New("runner cannot be empty")
	}
	if DetectRunner(c.Runner) == RunnerUnknown {
		return fmt.Errorf("unknown runner %q (supported runners: claude, cursor, pi, opencode, mock)", c.Runner)
	}
	return nil
}

func (c *Config) Validate() error {
	if err := c.ValidateRunner(); err != nil {
		return fmt.Errorf("invalid runner configuration: %w", err)
	}
	if c.PRDFile == "" {
		return errors.New("prd_file cannot be empty")
	}
	if c.TestCommand == "" {
		return errors.New("test_command cannot be empty")
	}

	// Prevent path traversal by requiring a simple filename.
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
