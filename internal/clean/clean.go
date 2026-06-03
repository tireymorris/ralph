package clean

import (
	"errors"
	"os"

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
	return removeIfExists(cfg.ConfigPath(workflow.ClarifyingQuestionsFile))
}

func removeIfExists(path string) error {
	err := os.Remove(path)
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
