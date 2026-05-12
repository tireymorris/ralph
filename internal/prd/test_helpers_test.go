package prd

import (
	"path/filepath"
	"testing"

	"ralph/internal/config"
)

func newTestConfig(t *testing.T, dir, file string) *config.Config {
	t.Helper()
	return &config.Config{PRDFile: filepath.Join(dir, file)}
}
