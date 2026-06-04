package clean

import (
	"os"
	"path/filepath"
	"time"

	"ralph/internal/shared/config"
)

func ArchivePriorState(cfg *config.Config) (backupDir string, err error) {
	prdPath := cfg.PRDPath()
	if _, err := os.Stat(prdPath); os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}
	backupDir, err = newBackupDir(cfg)
	if err != nil {
		return "", err
	}
	if err := moveIntoBackup(backupDir, prdPath, "prd.json"); err != nil {
		return "", err
	}
	return backupDir, nil
}

func newBackupDir(cfg *config.Config) (string, error) {
	ts := time.Now().UTC().Format("20060102T150405Z")
	dir := filepath.Join(cfg.WorkDir, ralphDataDir, "backups", ts)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func moveIntoBackup(backupDir, src, rel string) error {
	dest := filepath.Join(backupDir, rel)
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	return os.Rename(src, dest)
}
