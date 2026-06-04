package clean

import (
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/workflow"
)

func TestArchivePriorState_PRD(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig(t, dir)
	writeSeedFile(t, cfg.PRDPath())

	backupDir, err := ArchivePriorState(cfg)
	if err != nil {
		t.Fatalf("ArchivePriorState: %v", err)
	}
	if backupDir == "" {
		t.Fatal("backupDir empty, want timestamped backup path")
	}
	assertNotExist(t, cfg.PRDPath())
	if _, err := os.Stat(filepath.Join(backupDir, "prd.json")); err != nil {
		t.Fatalf("prd.json not in backup: %v", err)
	}
}

func TestArchivePriorState_stateFiles(t *testing.T) {
	tests := []struct {
		name string
		seed func(*config.Config) string
		rel  string
	}{
		{name: "PRD lock", seed: func(cfg *config.Config) string { return prd.LockPath(cfg.PRDPath()) }, rel: "prd.json.lock"},
		{
			name: "clarifying questions",
			seed: func(cfg *config.Config) string {
				return cfg.ConfigPath(workflow.ClarifyingQuestionsFile)
			},
			rel: ".ralph_questions.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			cfg := testConfig(t, dir)
			src := tt.seed(cfg)
			writeSeedFile(t, src)

			backupDir, err := ArchivePriorState(cfg)
			if err != nil {
				t.Fatalf("ArchivePriorState: %v", err)
			}
			if backupDir == "" {
				t.Fatal("backupDir empty")
			}
			assertNotExist(t, src)
			if _, err := os.Stat(filepath.Join(backupDir, tt.rel)); err != nil {
				t.Fatalf("%s not in backup: %v", tt.rel, err)
			}
		})
	}
}

func TestArchivePriorState_prdTemps(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig(t, dir)
	tmpPath := filepath.Join(dir, ".prd.tmp.100.7")
	writeSeedFile(t, tmpPath)

	backupDir, err := ArchivePriorState(cfg)
	if err != nil {
		t.Fatalf("ArchivePriorState: %v", err)
	}
	assertNoPRDTempFiles(t, dir)
	if _, err := os.Stat(filepath.Join(backupDir, ".prd.tmp.100.7")); err != nil {
		t.Fatalf("temp not in backup: %v", err)
	}
}

func TestArchivePriorState_runs(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig(t, dir)
	metaPath := filepath.Join(dir, ralphDataDir, "runs", "test-run", "meta.json")
	writeSeedFile(t, metaPath)

	backupDir, err := ArchivePriorState(cfg)
	if err != nil {
		t.Fatalf("ArchivePriorState: %v", err)
	}
	assertNotExist(t, metaPath)
	backedUp := filepath.Join(backupDir, "runs", "test-run", "meta.json")
	if _, err := os.Stat(backedUp); err != nil {
		t.Fatalf("run meta not in backup: %v", err)
	}
}

func TestArchivePriorState_allArtifacts(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig(t, dir)
	seeded, err := SeedStateArtifacts(cfg)
	if err != nil {
		t.Fatal(err)
	}
	priorBackup := filepath.Join(dir, ralphDataDir, "backups", "20200101T000000Z")
	if err := os.MkdirAll(priorBackup, 0755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(priorBackup, "keep.txt")
	if err := os.WriteFile(marker, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	backupDir, err := ArchivePriorState(cfg)
	if err != nil {
		t.Fatalf("ArchivePriorState: %v", err)
	}
	for _, p := range seeded {
		assertNotExist(t, p)
	}
	assertNotExist(t, cfg.PRDPath())
	assertNotExist(t, prd.LockPath(cfg.PRDPath()))
	assertNotExist(t, cfg.ConfigPath(workflow.ClarifyingQuestionsFile))
	matches, err := filepath.Glob(prdTempGlobPattern(cfg))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected no .prd.tmp.* files, got %v", matches)
	}
	runMeta := filepath.Join(dir, ralphDataDir, "runs", "test-run", "meta.json")
	assertNotExist(t, runMeta)
	if _, err := os.Stat(filepath.Join(backupDir, "runs", "test-run", "meta.json")); err != nil {
		t.Fatalf("run meta not in backup: %v", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("prior backup removed: %v", err)
	}
}

func TestArchivePriorState_noArtifacts(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig(t, dir)

	backupDir, err := ArchivePriorState(cfg)
	if err != nil {
		t.Fatalf("ArchivePriorState: %v", err)
	}
	if backupDir != "" {
		t.Fatalf("backupDir = %q, want empty", backupDir)
	}
	backupsRoot := filepath.Join(dir, ralphDataDir, "backups")
	if _, err := os.Stat(backupsRoot); !os.IsNotExist(err) {
		t.Fatalf(".ralph/backups should not exist: %v", err)
	}
}
