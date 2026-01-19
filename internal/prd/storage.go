package prd

import (
	"encoding/json"
	"fmt"
	"os"

	"ralph/internal/config"
)

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

func Delete(cfg *config.Config) error {
	if _, err := os.Stat(cfg.PRDFile); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(cfg.PRDFile)
}

func Exists(cfg *config.Config) bool {
	_, err := os.Stat(cfg.PRDFile)
	return err == nil
}
