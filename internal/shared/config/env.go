package config

import "os"

func applyEnvOverrides(cfg *Config) error {
	if runner := os.Getenv("RALPH_RUNNER"); runner != "" {
		cfg.Runner = runner
	}
	if os.Getenv("RALPH_YOLO") == "1" {
		cfg.AutoApprove = true
	}

	return cfg.Validate()
}
