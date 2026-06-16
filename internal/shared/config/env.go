package config

import (
	"fmt"
	"os"
	"strings"
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
	if prefix := os.Getenv("RALPH_BRANCH_PREFIX"); prefix != "" {
		cfg.BranchPrefix = prefix
	}
	if raw := os.Getenv("RALPH_DEFAULT_BRANCHES"); raw != "" {
		cfg.DefaultBranches = splitCommaList(raw)
	}
	if cmd := os.Getenv("RALPH_TEST_COMMAND"); cmd != "" {
		cfg.TestCommand = cmd
	}

	return cfg.Validate()
}

func splitCommaList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}
