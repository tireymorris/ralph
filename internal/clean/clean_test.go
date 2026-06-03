package clean

import (
	"os"
	"path/filepath"
	"testing"

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
	matches, err := filepath.Glob(filepath.Join(dir, ".prd.tmp.*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected no .prd.tmp.* files, got %v", matches)
	}
}

func TestRemoveState_seededArtifacts(t *testing.T) {
	tests := []struct {
		name        string
		seed        func(*config.Config) string
		assertExtra func(*testing.T, string, *config.Config)
	}{
		{
			name: "PRD",
			seed: func(cfg *config.Config) string { return cfg.PRDPath() },
		},
		{
			name: "PRD lock",
			seed: func(cfg *config.Config) string { return prd.LockPath(cfg.PRDPath()) },
		},
		{
			name: "clarifying questions",
			seed: func(cfg *config.Config) string {
				return cfg.ConfigPath(workflow.ClarifyingQuestionsFile)
			},
		},
		{
			name: "orphaned PRD temps",
			seed: func(cfg *config.Config) string {
				return filepath.Join(filepath.Dir(cfg.PRDPath()), ".prd.tmp.100.7")
			},
			assertExtra: func(t *testing.T, dir string, _ *config.Config) {
				assertNoPRDTempFiles(t, dir)
			},
		},
		{
			name: ".ralph run data",
			seed: func(cfg *config.Config) string {
				return filepath.Join(cfg.WorkDir, ralphDataDir, "runs", "x", "meta.json")
			},
			assertExtra: func(t *testing.T, _ string, cfg *config.Config) {
				assertNotExist(t, filepath.Join(cfg.WorkDir, ralphDataDir))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			cfg := testConfig(t, dir)
			writeSeedFile(t, tt.seed(cfg))
			if err := RemoveState(cfg); err != nil {
				t.Fatalf("RemoveState: %v", err)
			}
			if tt.assertExtra != nil {
				tt.assertExtra(t, dir, cfg)
			} else {
				assertNotExist(t, tt.seed(cfg))
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
