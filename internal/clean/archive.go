package clean

import (
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
)

func ArchivePriorState(cfg *config.Config) (backupDir string, err error) {
	has, err := hasStateArtifacts(cfg)
	if err != nil || !has {
		return "", err
	}
	releasePRDLock(cfg)
	backupDir, err = newBackupDir(cfg)
	if err != nil {
		return "", err
	}
	for _, path := range stateFilePaths(cfg) {
		if err := archiveFileIfExists(backupDir, path); err != nil {
			return "", err
		}
	}
	if err := archiveOrphanedPRDTemps(cfg, backupDir); err != nil {
		return "", err
	}
	return backupDir, nil
}

func archiveOrphanedPRDTemps(cfg *config.Config, backupDir string) error {
	matches, err := filepath.Glob(prdTempGlobPattern(cfg))
	if err != nil {
		return err
	}
	for _, path := range matches {
		if err := moveIntoBackup(backupDir, path, filepath.Base(path)); err != nil {
			return err
		}
	}
	return nil
}

func hasStateArtifacts(cfg *config.Config) (bool, error) {
	for _, path := range stateFilePaths(cfg) {
		if _, err := os.Stat(path); err == nil {
			return true, nil
		} else if !os.IsNotExist(err) {
			return false, err
		}
	}
	matches, err := filepath.Glob(prdTempGlobPattern(cfg))
	if err != nil {
		return false, err
	}
	if len(matches) > 0 {
		return true, nil
	}
	runsDir := filepath.Join(cfg.WorkDir, ralphDataDir, "runs")
	if _, err := os.Stat(runsDir); err == nil {
		return true, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	return false, nil
}

func releasePRDLock(cfg *config.Config) {
	lockPath := prd.LockPath(cfg.PRDPath())
	fileLock := flock.New(lockPath)
	_ = fileLock.Unlock()
}

func archiveFileIfExists(backupDir, src string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return moveIntoBackup(backupDir, src, filepath.Base(src))
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
