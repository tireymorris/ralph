package clean

import (
	"os"
	"path/filepath"
	"testing"
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
