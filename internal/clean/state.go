package clean

import (
	"os"
	"path/filepath"

	"ralph/internal/prompt"
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
		cfg.ConfigPath(prompt.PRDSelfReviewVerdictFile),
		// legacy root-level locations from before state moved under .ralph/
		cfg.ConfigPath(".ralph_questions.json"),
		cfg.ConfigPath(".ralph_prd_review.json"),
	}
}

func prdTempGlobPatterns(cfg *config.Config) []string {
	return []string{
		filepath.Join(cfg.WorkDir, ralphDataDir, "prd.tmp.*"),
		// legacy location next to prd.json from before state moved under .ralph/
		filepath.Join(filepath.Dir(cfg.PRDPath()), ".prd.tmp.*"),
	}
}

func runsDir(cfg *config.Config) string {
	return filepath.Join(cfg.WorkDir, ralphDataDir, "runs")
}

// SeedStateArtifacts creates all known Ralph state artifacts under cfg.WorkDir.
// Returned paths are the seeded files (not the .ralph directory itself).
func SeedStateArtifacts(cfg *config.Config) ([]string, error) {
	tmpPath := filepath.Join(cfg.WorkDir, ralphDataDir, "prd.tmp.1.999")
	legacyTmpPath := filepath.Join(filepath.Dir(cfg.PRDPath()), ".prd.tmp.1.999")
	metaPath := filepath.Join(runsDir(cfg), "test-run", "meta.json")
	paths := append(stateFilePaths(cfg), tmpPath, legacyTmpPath, metaPath)
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
