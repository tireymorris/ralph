package config

import "os"

func applyEnvOverrides(cfg *Config) error {
	if model := os.Getenv("RALPH_MODEL"); model != "" {
		cfg.Model = model
	}

	return cfg.Validate()
}
