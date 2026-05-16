package config

import "os"

func applyEnvOverrides(cfg *Config) error {
	if runner := os.Getenv("RALPH_RUNNER"); runner != "" {
		cfg.Runner = runner
	}

	return cfg.Validate()
}
