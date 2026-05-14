package review

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
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

func TestRunWithExistingPRDFile(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"

	// Write a minimal PRD file
	prdPath := filepath.Join(tmpDir, "prd.json")
	err := writeTestPRD(prdPath)
	if err != nil {
		t.Fatalf("failed to write test PRD: %v", err)
	}

	cmd := NewCmd(cfg, false)
	code := cmd.Run()

	if code != 0 {
		t.Fatalf("Run() = %d, want 0", code)
	}
}

func writeTestPRD(path string) error {
	p := &prd.PRD{
		Version:     1,
		ProjectName: "Test Project",
		Stories: []*prd.Story{
			{ID: "story-1", Title: "Test Story", Description: "A test story", AcceptanceCriteria: []string{"It works"}, Priority: 1},
		},
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
