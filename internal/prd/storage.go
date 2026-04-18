package prd

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ralph/internal/config"
	"ralph/internal/constants"
	"ralph/internal/logger"
)

// Load reads and parses the PRD from disk with a shared lock to prevent concurrent modifications.
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

	var p PRD
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse PRD file %q: %w", prdPath, err)
	}

	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("PRD validation failed for %q: %w", prdPath, err)
	}

	return &p, nil
}

// Save writes the PRD to disk using atomic file operations and exclusive locking.
// It acquires an exclusive lock, increments the version, writes to a temporary file,
// then atomically renames it to the final location. This ensures that:
// 1. The PRD file is never in a partially-written state (atomic writes)
// 2. No concurrent reads or writes occur during the save operation (exclusive lock)
// 3. Concurrent modifications can be detected via version numbers (optimistic locking)
func Save(cfg *config.Config, p *PRD) error {
	prdPath := cfg.PRDPath()

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

	dir := filepath.Dir(prdPath)

	rand.Seed(time.Now().UnixNano())
	tmpPath := filepath.Join(dir, fmt.Sprintf(".prd.tmp.%d.%d", time.Now().Unix(), rand.Intn(constants.TempFileRandomRange)))

	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write temporary PRD file %q: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, prdPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to atomically replace PRD file %q from temporary %q: %w", prdPath, tmpPath, err)
	}

	return nil
}

func Delete(cfg *config.Config) error {
	prdPath := cfg.PRDPath()
	if _, err := os.Stat(prdPath); os.IsNotExist(err) {
		return nil
	}
	if err := os.Remove(prdPath); err != nil {
		return fmt.Errorf("failed to delete PRD file %q: %w", prdPath, err)
	}
	return nil
}

func Exists(cfg *config.Config) bool {
	_, err := os.Stat(cfg.PRDPath())
	return err == nil
}

func Archive(cfg *config.Config) error {
	prdPath := cfg.PRDPath()
	if _, err := os.Stat(prdPath); os.IsNotExist(err) {
		return fmt.Errorf("PRD file %q does not exist", prdPath)
	}

	timestamp := time.Now().Format("20060102-150405")
	dir := filepath.Dir(prdPath)
	basename := filepath.Base(prdPath)
	ext := filepath.Ext(basename)
	name := strings.TrimSuffix(basename, ext)
	archiveName := fmt.Sprintf("%s-completed-%s%s", name, timestamp, ext)
	archivePath := filepath.Join(dir, archiveName)

	if err := os.Rename(prdPath, archivePath); err != nil {
		return fmt.Errorf("failed to archive PRD file to %q: %w", archivePath, err)
	}

	logger.Info("PRD archived", "path", archivePath)
	return nil
}
