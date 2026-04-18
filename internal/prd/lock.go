package prd

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/flock"
	"ralph/internal/config"
	"ralph/internal/constants"
)

// LockTimeoutError is returned when a file lock cannot be acquired within the timeout period.
type LockTimeoutError struct {
	Path    string
	Timeout time.Duration
}

func (e *LockTimeoutError) Error() string {
	return fmt.Sprintf("timeout acquiring lock on %s after %v", e.Path, e.Timeout)
}

// VersionConflictError is returned when the PRD version has changed unexpectedly,
// indicating concurrent modification.
type VersionConflictError struct {
	Expected int64
	Actual   int64
}

func (e *VersionConflictError) Error() string {
	return fmt.Sprintf("PRD version conflict: expected %d, got %d (concurrent modification detected)", e.Expected, e.Actual)
}

func getLockPath(prdPath string) string {
	return prdPath + ".lock"
}

func acquireSharedLock(cfg *config.Config) (*flock.Flock, error) {
	lockPath := getLockPath(cfg.PRDPath())
	fileLock := flock.New(lockPath)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(constants.FileLockTimeout)*time.Second)
	defer cancel()

	locked, err := fileLock.TryLockContext(ctx, time.Duration(constants.FileLockRetryDelay)*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("error acquiring shared lock: %w", err)
	}
	if !locked {
		return nil, &LockTimeoutError{Path: lockPath, Timeout: time.Duration(constants.FileLockTimeout) * time.Second}
	}

	return fileLock, nil
}

func acquireExclusiveLock(cfg *config.Config) (*flock.Flock, error) {
	lockPath := getLockPath(cfg.PRDPath())
	fileLock := flock.New(lockPath)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(constants.FileLockTimeout)*time.Second)
	defer cancel()

	locked, err := fileLock.TryLockContext(ctx, time.Duration(constants.FileLockRetryDelay)*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("error acquiring exclusive lock: %w", err)
	}
	if !locked {
		return nil, &LockTimeoutError{Path: lockPath, Timeout: time.Duration(constants.FileLockTimeout) * time.Second}
	}

	return fileLock, nil
}
