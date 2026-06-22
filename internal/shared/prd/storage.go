package prd

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/constants"
)

// Load reads and parses the PRD under a shared lock.
func Load(cfg *config.Config) (*PRD, error) {
	prdPath := cfg.PRDPath()

	fileLock, err := acquireSharedLock(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock for reading %q: %w", prdPath, err)
	}
	defer fileLock.Unlock()

	data, err := os.ReadFile(prdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PRD file %q: %w", prdPath, err)
	}

	if err := rejectLegacyAcceptanceCriteriaInJSON(data); err != nil {
		return nil, fmt.Errorf("PRD validation failed for %q: %w", prdPath, err)
	}

	var p PRD
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse PRD file %q: %w", prdPath, err)
	}

	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("PRD validation failed for %q: %w", prdPath, err)
	}

	return &p, nil
}

// Save writes the PRD atomically under an exclusive lock and increments Version.
func Save(cfg *config.Config, p *PRD) error {
	prdPath := cfg.PRDPath()

	if err := p.rejectLegacyAcceptanceCriteria(); err != nil {
		return fmt.Errorf("PRD validation failed before saving %q: %w", prdPath, err)
	}
	if err := p.Validate(); err != nil {
		return fmt.Errorf("PRD validation failed before saving %q: %w", prdPath, err)
	}

	fileLock, err := acquireExclusiveLock(cfg)
	if err != nil {
		return fmt.Errorf("failed to acquire lock for writing %q: %w", prdPath, err)
	}
	defer fileLock.Unlock()

	p.Version++

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal PRD %q (version %d): %w", prdPath, p.Version, err)
	}

	tmpDir := filepath.Join(filepath.Dir(prdPath), ".ralph")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("failed to create state dir %q: %w", tmpDir, err)
	}

	tmpPath := filepath.Join(tmpDir, fmt.Sprintf("prd.tmp.%d.%d", time.Now().Unix(), rand.Intn(constants.TempFileRandomRange)))

	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write temporary PRD file %q: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, prdPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to atomically replace PRD file %q from temporary %q: %w", prdPath, tmpPath, err)
	}

	return nil
}

func Exists(cfg *config.Config) (bool, error) {
	_, err := os.Stat(cfg.PRDPath())
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to stat PRD file %q: %w", cfg.PRDPath(), err)
}
