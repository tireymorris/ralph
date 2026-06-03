package clean

import (
	"errors"
	"os"
	"path/filepath"

	"ralph/internal/shared/config"
	"ralph/internal/workflow"
)

func RemoveState(cfg *config.Config) error {
	if err := removeIfExists(cfg.PRDPath()); err != nil {
		return err
	}
	if err := removeIfExists(cfg.PRDPath() + ".lock"); err != nil {
		return err
	}
	if err := removeIfExists(cfg.ConfigPath(workflow.ClarifyingQuestionsFile)); err != nil {
		return err
	}
	if err := removeOrphanedPRDTemps(cfg); err != nil {
		return err
	}
	return removeRalphDir(cfg)
}

func removeRalphDir(cfg *config.Config) error {
	err := os.RemoveAll(cfg.ConfigPath(".ralph"))
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func removeOrphanedPRDTemps(cfg *config.Config) error {
	pattern := filepath.Join(filepath.Dir(cfg.PRDPath()), ".prd.tmp.*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	for _, path := range matches {
		if err := removeIfExists(path); err != nil {
			return err
		}
	}
	return nil
}

func removeIfExists(path string) error {
	err := os.Remove(path)
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
