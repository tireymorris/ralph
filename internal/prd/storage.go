package prd

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
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

// getLockPath returns the path to the lock file for a given PRD path.
func getLockPath(prdPath string) string {
	return prdPath + ".lock"
}

// acquireSharedLock acquires a shared (read) lock on the PRD file.
// Multiple readers can hold shared locks simultaneously.
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

// acquireExclusiveLock acquires an exclusive (write) lock on the PRD file.
// Only one writer can hold the exclusive lock, blocking all other readers and writers.
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

// Load reads and parses the PRD from disk with a shared lock to prevent concurrent modifications.
func Load(cfg *config.Config) (*PRD, error) {
	prdPath := cfg.PRDPath()

	// Acquire shared lock (allows multiple concurrent readers)
	fileLock, err := acquireSharedLock(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock for reading %q: %w", prdPath, err)
	}
	defer fileLock.Unlock()

	data, err := os.ReadFile(prdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PRD file %q: %w", prdPath, err)
	}

	var p PRD
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse PRD file %q: %w", prdPath, err)
	}

	// Validate loaded PRD data
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("PRD validation failed for %q: %w", prdPath, err)
	}

	return &p, nil
}

// Save writes the PRD to disk using atomic file operations and exclusive locking.
// It acquires an exclusive lock, increments the version, writes to a temporary file,
// then atomically renames it to the final location. This ensures that:
// 1. The PRD file is never in a partially-written state (atomic writes)
// 2. No concurrent reads or writes occur during the save operation (exclusive lock)
// 3. Concurrent modifications can be detected via version numbers (optimistic locking)
func Save(cfg *config.Config, p *PRD) error {
	prdPath := cfg.PRDPath()

	// Validate PRD before saving
	if err := p.Validate(); err != nil {
		return fmt.Errorf("PRD validation failed before saving %q: %w", prdPath, err)
	}

	// Acquire exclusive lock (blocks all readers and other writers)
	fileLock, err := acquireExclusiveLock(cfg)
	if err != nil {
		return fmt.Errorf("failed to acquire lock for writing %q: %w", prdPath, err)
	}
	defer fileLock.Unlock()

	// Increment version for optimistic locking detection
	p.Version++

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal PRD %q (version %d): %w", prdPath, p.Version, err)
	}

	dir := filepath.Dir(prdPath)

	// Create a temporary file in the same directory as the target file.
	// This ensures that the rename operation will be atomic (same filesystem).
	rand.Seed(time.Now().UnixNano())
	tmpPath := filepath.Join(dir, fmt.Sprintf(".prd.tmp.%d.%d", time.Now().Unix(), rand.Intn(constants.TempFileRandomRange)))

	// Write to temp file with restricted permissions (user-only read/write)
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write temporary PRD file %q: %w", tmpPath, err)
	}

	// Atomic rename: this is atomic on Unix/Linux/macOS as long as both files
	// are on the same filesystem (which they are, since we used the same dir).
	if err := os.Rename(tmpPath, prdPath); err != nil {
		// Clean up temp file on error
		if removeErr := os.Remove(tmpPath); removeErr != nil {
			// Log cleanup error but don't override original error
		}
		return fmt.Errorf("failed to atomically replace PRD file %q from temporary %q: %w", prdPath, tmpPath, err)
	}

	return nil
}

func Delete(cfg *config.Config) error {
	prdPath := cfg.PRDPath()
	if _, err := os.Stat(prdPath); os.IsNotExist(err) {
		return nil
	}
	if err := os.Remove(prdPath); err != nil {
		return fmt.Errorf("failed to delete PRD file %q: %w", prdPath, err)
	}
	return nil
}

func Exists(cfg *config.Config) bool {
	_, err := os.Stat(cfg.PRDPath())
	return err == nil
}
