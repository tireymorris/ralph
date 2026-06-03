package clean

import (
	"os"
	"path/filepath"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/workflow"
)

const ralphDataDir = ".ralph"

func stateFilePaths(cfg *config.Config) []string {
	return []string{
		cfg.PRDPath(),
		prd.LockPath(cfg.PRDPath()),
		cfg.ConfigPath(workflow.ClarifyingQuestionsFile),
	}
}

func prdTempGlobPattern(cfg *config.Config) string {
	return filepath.Join(filepath.Dir(cfg.PRDPath()), ".prd.tmp.*")
}

// SeedStateArtifacts creates all known Ralph state artifacts under cfg.WorkDir.
// Returned paths are the seeded files (not the .ralph directory itself).
func SeedStateArtifacts(cfg *config.Config) ([]string, error) {
	tmpPath := filepath.Join(filepath.Dir(cfg.PRDPath()), ".prd.tmp.1.999")
	metaPath := filepath.Join(cfg.WorkDir, ralphDataDir, "runs", "test-run", "meta.json")
	paths := append(stateFilePaths(cfg), tmpPath, metaPath)
	for _, p := range paths {
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(p, []byte("{}"), 0644); err != nil {
			return nil, err
		}
	}
	return paths, nil
}
