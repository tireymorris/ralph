package config

import (
	"fmt"
	"os"
	"strconv"
)

func applyEnvOverrides(cfg *Config) error {
	if model := os.Getenv("RALPH_MODEL"); model != "" {
		cfg.Model = model
	}

	if maxIterStr := os.Getenv("RALPH_MAX_ITERATIONS"); maxIterStr != "" {
		maxIter, err := strconv.Atoi(maxIterStr)
		if err != nil {
			return fmt.Errorf("invalid RALPH_MAX_ITERATIONS value %q: %w", maxIterStr, err)
		}
		if maxIter > 0 {
			cfg.MaxIterations = maxIter
		}
	}

	if retryStr := os.Getenv("RALPH_RETRY_ATTEMPTS"); retryStr != "" {
		retry, err := strconv.Atoi(retryStr)
		if err != nil {
			return fmt.Errorf("invalid RALPH_RETRY_ATTEMPTS value %q: %w", retryStr, err)
		}
		if retry >= 0 {
			cfg.RetryAttempts = retry
		}
	}

	if prdFile := os.Getenv("RALPH_PRD_FILE"); prdFile != "" {
		cfg.PRDFile = prdFile
	}

	if testCmd := os.Getenv("RALPH_TEST_COMMAND"); testCmd != "" {
		cfg.TestCommand = testCmd
	}

	return cfg.Validate()
}
