package prd

import (
	"encoding/json"
	"fmt"
	"os"

	"ralph/internal/config"
)

// Load reads a PRD from the configured file
func Load(cfg *config.Config) (*PRD, error) {
	data, err := os.ReadFile(cfg.PRDFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read PRD file: %w", err)
	}

	var p PRD
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse PRD: %w", err)
	}

	return &p, nil
}

// Save writes the PRD to the configured file
func Save(cfg *config.Config, p *PRD) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal PRD: %w", err)
	}

	if err := os.WriteFile(cfg.PRDFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write PRD file: %w", err)
	}

	return nil
}

// Delete removes the PRD file
func Delete(cfg *config.Config) error {
	if _, err := os.Stat(cfg.PRDFile); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(cfg.PRDFile)
}

// Exists checks if a PRD file exists
func Exists(cfg *config.Config) bool {
	_, err := os.Stat(cfg.PRDFile)
	return err == nil
}
