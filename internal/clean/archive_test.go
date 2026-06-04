package clean

import (
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/shared/prd"
	"ralph/internal/workflow"
)

func TestArchivePriorState_seededArtifacts(t *testing.T) {
	for _, tt := range stateArtifactCases() {
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
			if _, err := os.Stat(filepath.Join(backupDir, tt.backupRel)); err != nil {
				t.Fatalf("%s not in backup: %v", tt.backupRel, err)
			}
		})
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
	assertNoPRDTempFiles(t, dir)
	runMeta := filepath.Join(runsDir(cfg), "test-run", "meta.json")
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
