package clean

import (
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/workflow"
)

func testConfig(t *testing.T, dir string) *config.Config {
	t.Helper()
	return &config.Config{WorkDir: dir, PRDFile: "prd.json"}
}

func writeSeedFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
}

func assertNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("%s still exists: %v", path, err)
	}
}

func assertNoPRDTempFiles(t *testing.T, dir string) {
	t.Helper()
	for _, pattern := range []string{
		filepath.Join(dir, ralphDataDir, "prd.tmp.*"),
		filepath.Join(dir, ".prd.tmp.*"),
	} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) != 0 {
			t.Fatalf("expected no PRD temp files, got %v", matches)
		}
	}
}

type stateArtifactCase struct {
	name         string
	seed         func(*config.Config) string
	backupRel    string
	removeAssert func(t *testing.T, dir string, cfg *config.Config, seedPath string)
}

func stateArtifactCases() []stateArtifactCase {
	return []stateArtifactCase{
		{
			name:      "PRD",
			seed:      func(cfg *config.Config) string { return cfg.PRDPath() },
			backupRel: "prd.json",
		},
		{
			name:      "PRD lock",
			seed:      func(cfg *config.Config) string { return prd.LockPath(cfg.PRDPath()) },
			backupRel: "prd.json.lock",
		},
		{
			name: "clarifying questions",
			seed: func(cfg *config.Config) string {
				return cfg.ConfigPath(workflow.ClarifyingQuestionsFile)
			},
			backupRel: "questions.json",
		},
		{
			name: "self-review verdict",
			seed: func(cfg *config.Config) string {
				return cfg.ConfigPath(prompt.PRDSelfReviewVerdictFile)
			},
			backupRel: "prd_review.json",
		},
		{
			name: "orphaned PRD temps",
			seed: func(cfg *config.Config) string {
				return filepath.Join(cfg.WorkDir, ralphDataDir, "prd.tmp.100.7")
			},
			backupRel: "prd.tmp.100.7",
			removeAssert: func(t *testing.T, dir string, _ *config.Config, _ string) {
				assertNoPRDTempFiles(t, dir)
			},
		},
		{
			name: "legacy orphaned PRD temps",
			seed: func(cfg *config.Config) string {
				return filepath.Join(filepath.Dir(cfg.PRDPath()), ".prd.tmp.100.7")
			},
			backupRel: ".prd.tmp.100.7",
			removeAssert: func(t *testing.T, dir string, _ *config.Config, _ string) {
				assertNoPRDTempFiles(t, dir)
			},
		},
		{
			name: "runs",
			seed: func(cfg *config.Config) string {
				return filepath.Join(runsDir(cfg), "test-run", "meta.json")
			},
			backupRel: "runs/test-run/meta.json",
			removeAssert: func(t *testing.T, _ string, cfg *config.Config, _ string) {
				assertNotExist(t, filepath.Join(cfg.WorkDir, ralphDataDir))
			},
		},
	}
}

func TestCleanRemovesLegacyClarifyingQuestions(t *testing.T) {
	cfg := testConfig(t, t.TempDir())
	for _, legacy := range []string{".ralph_questions.json", ".ralph_prd_review.json"} {
		path := cfg.ConfigPath(legacy)
		for _, p := range stateFilePaths(cfg) {
			if p == path {
				t.Fatalf("stateFilePaths still includes legacy path %s", legacy)
			}
		}
	}
}

func TestRemoveState_seededArtifacts(t *testing.T) {
	for _, tt := range stateArtifactCases() {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			cfg := testConfig(t, dir)
			seedPath := tt.seed(cfg)
			writeSeedFile(t, seedPath)
			if err := RemoveState(cfg); err != nil {
				t.Fatalf("RemoveState: %v", err)
			}
			if tt.removeAssert != nil {
				tt.removeAssert(t, dir, cfg, seedPath)
			} else {
				assertNotExist(t, seedPath)
			}
		})
	}
}

func TestRemoveState_allArtifacts(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig(t, dir)
	seeded, err := SeedStateArtifacts(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if err := RemoveState(cfg); err != nil {
		t.Fatalf("RemoveState: %v", err)
	}
	for _, p := range seeded {
		assertNotExist(t, p)
	}
	assertNotExist(t, filepath.Join(dir, ralphDataDir))
	assertNoPRDTempFiles(t, dir)
}

func TestRemoveState_idempotent(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig(t, dir)
	if _, err := SeedStateArtifacts(cfg); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 2; i++ {
		if err := RemoveState(cfg); err != nil {
			t.Fatalf("RemoveState call %d: %v", i, err)
		}
	}
}

func TestRemoveState_idempotentOnEmptyDir(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig(t, dir)
	for i := 1; i <= 2; i++ {
		if err := RemoveState(cfg); err != nil {
			t.Fatalf("RemoveState call %d: %v", i, err)
		}
	}
}
