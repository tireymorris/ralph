package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/shared/config"
)

func TestMockRunnerWritesPRD(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.Runner = "mock"
	cfg.PRDFile = "prd.json"

	r := NewMock(cfg)
	ch := make(chan OutputLine, 4)
	if err := r.Run(context.Background(), "Generate a PRD.\n3. Write the PRD file, then STOP.", ch); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, "prd.json")); err != nil {
		t.Fatalf("prd.json not written: %v", err)
	}
}
