package prd

import (
	"encoding/json"
	"fmt"
	"os"

	"ralph/internal/config"
)

// Load reads a PRD from the file system.
func Load(cfg *config.Config) (*PRD, error) {
	data, err := os.ReadFile(cfg.PRDPath())
	if err != nil {
		return nil, fmt.Errorf("failed to read PRD file: %w", err)
	}

	var p PRD
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse PRD: %w", err)
	}

	return &p, nil
}

// Save writes a PRD to the file system.
func Save(cfg *config.Config, p *PRD) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal PRD: %w", err)
	}

	if err := os.WriteFile(cfg.PRDPath(), data, 0644); err != nil {
		return fmt.Errorf("failed to write PRD file: %w", err)
	}

	return nil
}

// Delete removes the PRD file from the file system.
func Delete(cfg *config.Config) error {
	prdPath := cfg.PRDPath()
	if _, err := os.Stat(prdPath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(prdPath)
}

// Exists checks if the PRD file exists on the file system.
func Exists(cfg *config.Config) bool {
	_, err := os.Stat(cfg.PRDPath())
	return err == nil
}
