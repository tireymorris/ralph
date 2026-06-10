package config

import (
	"fmt"
	"os"
	"time"
)

func applyEnvOverrides(cfg *Config) error {
	if runner := os.Getenv("RALPH_RUNNER"); runner != "" {
		cfg.Runner = runner
	}
	if os.Getenv("RALPH_YOLO") == "1" {
		cfg.AutoApprove = true
	}
	if rawTimeout := os.Getenv("RALPH_RUNNER_TIMEOUT"); rawTimeout != "" {
		timeout, err := time.ParseDuration(rawTimeout)
		if err != nil {
			return fmt.Errorf("RALPH_RUNNER_TIMEOUT must be a Go duration: %w", err)
		}
		cfg.RunnerTimeout = timeout
	}

	return cfg.Validate()
}
