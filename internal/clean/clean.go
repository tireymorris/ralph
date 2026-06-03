package clean

import (
	"errors"
	"os"

	"ralph/internal/shared/config"
)

func RemoveState(cfg *config.Config) error {
	return removeIfExists(cfg.PRDPath())
}

func removeIfExists(path string) error {
	err := os.Remove(path)
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
