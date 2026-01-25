package prd

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"ralph/internal/config"
)

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	prdFile := filepath.Join(tmpDir, "test.json")
	cfg := &config.Config{PRDFile: prdFile}

	original := &PRD{
		ProjectName: "Test Project",
		BranchName:  "feature/test",
		Stories: []*Story{
			{
				ID:                 "story-1",
				Title:              "Test Story",
				Description:        "A test story",
				AcceptanceCriteria: []string{"criterion 1"},
				TestSpec:           "test spec",
				Priority:           1,
				Passes:             false,
				RetryCount:         0,
			},
		},
	}

	err := Save(cfg, original)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(cfg)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.ProjectName != original.ProjectName {
		t.Errorf("ProjectName = %q, want %q", loaded.ProjectName, original.ProjectName)
	}
	if loaded.BranchName != original.BranchName {
		t.Errorf("BranchName = %q, want %q", loaded.BranchName, original.BranchName)
	}
	if len(loaded.Stories) != len(original.Stories) {
		t.Errorf("Stories count = %d, want %d", len(loaded.Stories), len(original.Stories))
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	cfg := &config.Config{PRDFile: "/nonexistent/path/file.json"}

	_, err := Load(cfg)
	if err == nil {
		t.Error("Load() expected error for non-existent file")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	prdFile := filepath.Join(tmpDir, "invalid.json")
	cfg := &config.Config{PRDFile: prdFile}

	os.WriteFile(prdFile, []byte("not valid json"), 0644)

	_, err := Load(cfg)
	if err == nil {
		t.Error("Load() expected error for invalid JSON")
	}
}

func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()
	prdFile := filepath.Join(tmpDir, "delete.json")
	cfg := &config.Config{PRDFile: prdFile}

	os.WriteFile(prdFile, []byte("{}"), 0644)

	if !Exists(cfg) {
		t.Error("file should exist before delete")
	}

	err := Delete(cfg)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if Exists(cfg) {
		t.Error("file should not exist after delete")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	cfg := &config.Config{PRDFile: "/nonexistent/file.json"}

	err := Delete(cfg)
	if err != nil {
		t.Errorf("Delete() should not error for non-existent file, got %v", err)
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("file exists", func(t *testing.T) {
		prdFile := filepath.Join(tmpDir, "exists.json")
		cfg := &config.Config{PRDFile: prdFile}

		os.WriteFile(prdFile, []byte("{}"), 0644)

		if !Exists(cfg) {
			t.Error("Exists() = false, want true")
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		cfg := &config.Config{PRDFile: filepath.Join(tmpDir, "not-exists.json")}

		if Exists(cfg) {
			t.Error("Exists() = true, want false")
		}
	})
}

func TestSaveUnwritableLocation(t *testing.T) {
	cfg := &config.Config{PRDFile: "/nonexistent/dir/file.json"}
	prd := &PRD{ProjectName: "test"}

	err := Save(cfg, prd)
	if err == nil {
		t.Error("Save() expected error for unwritable location")
	}
}

// TestVersionIncrement verifies that Version increments on each save
func TestVersionIncrement(t *testing.T) {
	tmpDir := t.TempDir()
	prdFile := filepath.Join(tmpDir, "version.json")
	cfg := &config.Config{PRDFile: prdFile}

	prd := &PRD{
		ProjectName: "Test Project",
		Version:     0,
	}

	// Save multiple times and verify version increments
	for i := 1; i <= 5; i++ {
		if err := Save(cfg, prd); err != nil {
			t.Fatalf("Save %d failed: %v", i, err)
		}

		if prd.Version != int64(i) {
			t.Errorf("after save %d: expected version %d, got %d", i, i, prd.Version)
		}

		// Load and verify version persisted to disk
		loaded, err := Load(cfg)
		if err != nil {
			t.Fatalf("Load after save %d failed: %v", i, err)
		}

		if loaded.Version != int64(i) {
			t.Errorf("after save %d: expected loaded version %d, got %d", i, i, loaded.Version)
		}
	}
}

// TestBackwardsCompatibilityNoVersion verifies that PRDs without version field still load
func TestBackwardsCompatibilityNoVersion(t *testing.T) {
	tmpDir := t.TempDir()
	prdFile := filepath.Join(tmpDir, "old.json")
	cfg := &config.Config{PRDFile: prdFile}

	// Manually write a PRD JSON without version field (old format)
	oldFormatJSON := `{
  "project_name": "Old Project",
  "stories": [
    {
      "id": "story-1",
      "title": "Old Story",
      "description": "test",
      "acceptance_criteria": [],
      "priority": 1,
      "passes": false,
      "retry_count": 0
    }
  ]
}`

	if err := os.WriteFile(prdFile, []byte(oldFormatJSON), 0644); err != nil {
		t.Fatalf("failed to write old format PRD: %v", err)
	}

	// Load should succeed and default version to 0
	loaded, err := Load(cfg)
	if err != nil {
		t.Fatalf("Load failed for old format PRD: %v", err)
	}

	if loaded.Version != 0 {
		t.Errorf("expected version 0 for old format PRD, got %d", loaded.Version)
	}

	if loaded.ProjectName != "Old Project" {
		t.Errorf("expected project name %q, got %q", "Old Project", loaded.ProjectName)
	}
}

// TestAtomicWriteFilePermissions verifies file permissions are 0600
func TestAtomicWriteFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	prdFile := filepath.Join(tmpDir, "perms.json")
	cfg := &config.Config{PRDFile: prdFile}

	prd := &PRD{ProjectName: "Test"}

	if err := Save(cfg, prd); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	info, err := os.Stat(prdFile)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("expected file permissions 0600, got %o", mode)
	}
}

// TestAtomicWriteNoTempFiles verifies temp files are cleaned up
func TestAtomicWriteNoTempFiles(t *testing.T) {
	tmpDir := t.TempDir()
	prdFile := filepath.Join(tmpDir, "cleanup.json")
	cfg := &config.Config{PRDFile: prdFile}

	prd := &PRD{ProjectName: "Test"}

	if err := Save(cfg, prd); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Check for temp files
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read temp dir: %v", err)
	}

	for _, file := range files {
		name := file.Name()
		if filepath.Ext(name) == ".tmp" || name != filepath.Base(prdFile) && name != filepath.Base(prdFile)+".lock" {
			t.Errorf("unexpected file in temp dir: %s", name)
		}
	}
}

// TestConcurrentReads verifies multiple concurrent reads can succeed
func TestConcurrentReads(t *testing.T) {
	tmpDir := t.TempDir()
	prdFile := filepath.Join(tmpDir, "concurrent-read.json")
	cfg := &config.Config{PRDFile: prdFile}

	prd := &PRD{
		ProjectName: "Concurrent Test",
		Stories:     []*Story{{ID: "story-1", Title: "Test", Priority: 1}},
	}

	if err := Save(cfg, prd); err != nil {
		t.Fatalf("initial Save failed: %v", err)
	}

	// Launch multiple concurrent readers
	numReaders := 10
	var wg sync.WaitGroup
	wg.Add(numReaders)

	errors := make(chan error, numReaders)

	for i := 0; i < numReaders; i++ {
		go func() {
			defer wg.Done()

			loaded, err := Load(cfg)
			if err != nil {
				errors <- err
				return
			}

			if loaded.ProjectName != "Concurrent Test" {
				t.Error("loaded wrong data")
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Errorf("concurrent read failed: %v", err)
		}
	}
}

// TestConcurrentWrites verifies concurrent writes are serialized correctly
func TestConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	prdFile := filepath.Join(tmpDir, "concurrent-write.json")
	cfg := &config.Config{PRDFile: prdFile}

	prd := &PRD{
		ProjectName: "Write Test",
		Version:     0,
	}

	if err := Save(cfg, prd); err != nil {
		t.Fatalf("initial Save failed: %v", err)
	}

	// Launch multiple concurrent writers
	numWriters := 10
	var wg sync.WaitGroup
	wg.Add(numWriters)

	errors := make(chan error, numWriters)

	for i := 0; i < numWriters; i++ {
		go func(id int) {
			defer wg.Done()

			current, err := Load(cfg)
			if err != nil {
				errors <- err
				return
			}

			current.ProjectName = "Modified"
			if err := Save(cfg, current); err != nil {
				errors <- err
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Errorf("concurrent write failed: %v", err)
		}
	}

	// Verify final version incremented (concurrent access is serialized)
	final, err := Load(cfg)
	if err != nil {
		t.Fatalf("final Load failed: %v", err)
	}

	// Version should have incremented, though exact final value depends on
	// scheduling of concurrent loads/saves. Just verify it increased.
	if final.Version <= 1 {
		t.Errorf("expected version > 1, got %d", final.Version)
	}

	// Verify no corruption
	if final.ProjectName != "Modified" {
		t.Error("final PRD has corrupted data")
	}
}

// TestLockTimeoutError verifies the LockTimeoutError type
func TestLockTimeoutError(t *testing.T) {
	err := &LockTimeoutError{
		Path:    "/tmp/test.lock",
		Timeout: 30 * time.Second,
	}

	expected := "timeout acquiring lock on /tmp/test.lock after 30s"
	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

// TestVersionConflictError verifies the VersionConflictError type
func TestVersionConflictError(t *testing.T) {
	err := &VersionConflictError{
		Expected: 5,
		Actual:   10,
	}

	expected := "PRD version conflict: expected 5, got 10 (concurrent modification detected)"
	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

// BenchmarkSave measures performance of Save operation
func BenchmarkSave(b *testing.B) {
	tmpDir := b.TempDir()
	prdFile := filepath.Join(tmpDir, "bench.json")
	cfg := &config.Config{PRDFile: prdFile}

	prd := &PRD{
		ProjectName: "Benchmark",
		Stories:     []*Story{{ID: "story-1", Title: "Test", Priority: 1}},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := Save(cfg, prd); err != nil {
			b.Fatalf("Save failed: %v", err)
		}
	}
}

// BenchmarkLoad measures performance of Load operation
func BenchmarkLoad(b *testing.B) {
	tmpDir := b.TempDir()
	prdFile := filepath.Join(tmpDir, "bench.json")
	cfg := &config.Config{PRDFile: prdFile}

	prd := &PRD{
		ProjectName: "Benchmark",
		Stories:     []*Story{{ID: "story-1", Title: "Test", Priority: 1}},
	}

	Save(cfg, prd)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := Load(cfg); err != nil {
			b.Fatalf("Load failed: %v", err)
		}
	}
}
