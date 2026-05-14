package implement

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
)

func writeTestPRD(path string, stories []*prd.Story) error {
	p := &prd.PRD{
		Version:     1,
		ProjectName: "Test Project",
		Stories:     stories,
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

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

func TestRunWithAllCompleteExitsEarly(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"

	prdPath := filepath.Join(tmpDir, "prd.json")
	stories := []*prd.Story{
		{ID: "story-1", Title: "Done Story", Description: "Already done", AcceptanceCriteria: []string{"It works"}, Priority: 1, Passes: true},
	}
	err := writeTestPRD(prdPath, stories)
	if err != nil {
		t.Fatalf("failed to write test PRD: %v", err)
	}

	cmd := NewCmd(cfg, false)
	code := cmd.Run()

	if code != 0 {
		t.Fatalf("Run() = %d, want 0", code)
	}
}
