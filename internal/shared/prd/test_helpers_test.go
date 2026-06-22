package prd

import (
	"path/filepath"
	"testing"

	"ralph/internal/shared/config"
)

func newTestConfig(t *testing.T, dir, file string) *config.Config {
	t.Helper()
	return &config.Config{PRDFile: filepath.Join(dir, file)}
}

func testSlice(behavior string) []*Slice {
	return []*Slice{{
		ID:       "slice-1",
		Behavior: behavior,
		RedHint:  "add failing test",
	}}
}
