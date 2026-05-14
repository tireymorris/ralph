package implement

import (
	"testing"

	"ralph/internal/shared/config"
)

func TestRunWithNoPRDFile(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"

	cmd := NewCmd(cfg, false)
	code := cmd.Run()

	if code != 1 {
		t.Fatalf("Run() = %d, want 1", code)
	}
}
