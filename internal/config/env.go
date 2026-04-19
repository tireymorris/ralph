package config

import (
	"os"
)

func applyEnvOverrides(cfg *Config) error {
	if model := os.Getenv("RALPH_MODEL"); model != "" {
		cfg.Model = model
	}

	if prdFile := os.Getenv("RALPH_PRD_FILE"); prdFile != "" {
		cfg.PRDFile = prdFile
	}

	if testCmd := os.Getenv("RALPH_TEST_COMMAND"); testCmd != "" {
		cfg.TestCommand = testCmd
	}

	return cfg.Validate()
}
