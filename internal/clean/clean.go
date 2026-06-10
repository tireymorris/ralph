package clean

import (
	"errors"
	"os"
	"path/filepath"

	"ralph/internal/shared/config"
)

func RemoveState(cfg *config.Config) error {
	for _, path := range stateFilePaths(cfg) {
		if err := removeIfExists(path); err != nil {
			return err
		}
	}
	if err := removeOrphanedPRDTemps(cfg); err != nil {
		return err
	}
	return removeTree(cfg.ConfigPath(ralphDataDir))
}

func removeOrphanedPRDTemps(cfg *config.Config) error {
	for _, pattern := range prdTempGlobPatterns(cfg) {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return err
		}
		for _, path := range matches {
			if err := removeIfExists(path); err != nil {
				return err
			}
		}
	}
	return nil
}

func removeTree(path string) error {
	err := os.RemoveAll(path)
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func removeIfExists(path string) error {
	err := os.Remove(path)
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
