package prd

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"ralph/internal/shared/config"
)

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := newTestConfig(t, tmpDir, "test.json")

	original := &PRD{
		ProjectName: "Test Project",
		BranchName:  "feature/test",
		Stories: []*Story{
			{
				ID:          "story-1",
				Title:       "Test Story",
				Description: "A test story",
				Slices: []*Slice{
					{ID: "slice-1", Behavior: "criterion 1", RedHint: "add failing test"},
				},
				Priority: 1,
				Passes:   false,
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

	data, err := os.ReadFile(cfg.PRDPath())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(data), "acceptance_criteria") {
		t.Fatalf("saved PRD should not contain acceptance_criteria: %s", data)
	}
}

func TestLoadRejectsLegacyAcceptanceCriteria(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := newTestConfig(t, tmpDir, "legacy-reject.json")

	legacyJSON := `{
  "project_name": "Legacy Project",
  "stories": [
    {
      "id": "story-1",
      "title": "Legacy Story",
      "description": "A legacy story",
      "acceptance_criteria": ["criterion 1"],
      "priority": 1,
      "passes": false
    }
  ]
}`
	if err := os.WriteFile(cfg.PRDPath(), []byte(legacyJSON), 0600); err != nil {
		t.Fatalf("write legacy PRD: %v", err)
	}

	_, err := Load(cfg)
	if err == nil {
		t.Fatal("Load() expected error for legacy acceptance_criteria")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "acceptance") {
		t.Fatalf("Load() error = %q, want message mentioning acceptance criteria", err)
	}
}

func TestLoadRejectsMixedAcceptanceCriteriaAndSlices(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := newTestConfig(t, tmpDir, "mixed.json")

	mixedJSON := `{
  "project_name": "Mixed Project",
  "stories": [
    {
      "id": "story-1",
      "title": "Mixed Story",
      "description": "Has both formats",
      "acceptance_criteria": ["criterion 1"],
      "slices": [{"id": "slice-1", "behavior": "b", "red_hint": "r"}],
      "priority": 1,
      "passes": false
    }
  ]
}`
	if err := os.WriteFile(cfg.PRDPath(), []byte(mixedJSON), 0600); err != nil {
		t.Fatalf("write mixed PRD: %v", err)
	}

	_, err := Load(cfg)
	if err == nil {
		t.Fatal("Load() expected error for mixed acceptance_criteria and slices")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "acceptance") {
		t.Fatalf("Load() error = %q, want message mentioning acceptance criteria", err)
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
	cfg := newTestConfig(t, tmpDir, "invalid.json")

	os.WriteFile(cfg.PRDPath(), []byte("not valid json"), 0644)

	_, err := Load(cfg)
	if err == nil {
		t.Error("Load() expected error for invalid JSON")
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("file exists", func(t *testing.T) {
		cfg := newTestConfig(t, tmpDir, "exists.json")

		os.WriteFile(cfg.PRDPath(), []byte("{}"), 0644)

		exists, err := Exists(cfg)
		if err != nil {
			t.Fatalf("Exists() error = %v", err)
		}
		if !exists {
			t.Error("Exists() = false, want true")
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		cfg := newTestConfig(t, tmpDir, "not-exists.json")

		exists, err := Exists(cfg)
		if err != nil {
			t.Fatalf("Exists() error = %v", err)
		}
		if exists {
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

func TestVersionIncrement(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := newTestConfig(t, tmpDir, "version.json")

	prd := &PRD{
		ProjectName: "Test Project",
		Version:     0,
	}

	for i := 1; i <= 5; i++ {
		if err := Save(cfg, prd); err != nil {
			t.Fatalf("Save %d failed: %v", i, err)
		}

		if prd.Version != int64(i) {
			t.Errorf("after save %d: expected version %d, got %d", i, i, prd.Version)
		}

		loaded, err := Load(cfg)
		if err != nil {
			t.Fatalf("Load after save %d failed: %v", i, err)
		}

		if loaded.Version != int64(i) {
			t.Errorf("after save %d: expected loaded version %d, got %d", i, i, loaded.Version)
		}
	}
}

func TestLoadRejectsEmptyAcceptanceCriteriaWithoutSlices(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := newTestConfig(t, tmpDir, "old.json")

	oldFormatJSON := `{
  "project_name": "Old Project",
  "stories": [
    {
      "id": "story-1",
      "title": "Old Story",
      "description": "test",
      "acceptance_criteria": [],
      "priority": 1,
      "passes": false
    }
  ]
}`

	if err := os.WriteFile(cfg.PRDPath(), []byte(oldFormatJSON), 0644); err != nil {
		t.Fatalf("failed to write old format PRD: %v", err)
	}

	_, err := Load(cfg)
	if err == nil {
		t.Fatal("Load() expected error for empty acceptance_criteria without slices")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "acceptance") {
		t.Fatalf("Load() error = %q, want message mentioning acceptance criteria", err)
	}
}

func TestAtomicWriteFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := newTestConfig(t, tmpDir, "perms.json")

	prd := &PRD{ProjectName: "Test"}

	if err := Save(cfg, prd); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	info, err := os.Stat(cfg.PRDPath())
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("expected file permissions 0600, got %o", mode)
	}
}

func TestAtomicWriteNoTempFiles(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := newTestConfig(t, tmpDir, "cleanup.json")

	prd := &PRD{ProjectName: "Test"}

	if err := Save(cfg, prd); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read temp dir: %v", err)
	}

	for _, file := range files {
		name := file.Name()
		base := filepath.Base(cfg.PRDFile)
		if name == ".ralph" {
			continue
		}
		if filepath.Ext(name) == ".tmp" || name != base && name != base+".lock" {
			t.Errorf("unexpected file in temp dir: %s", name)
		}
	}
}

func TestConcurrentReads(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := newTestConfig(t, tmpDir, "concurrent-read.json")

	prd := &PRD{
		ProjectName: "Concurrent Test",
		Stories:     []*Story{{ID: "story-1", Title: "Test", Priority: 1, Slices: testSlice("works")}},
	}

	if err := Save(cfg, prd); err != nil {
		t.Fatalf("initial Save failed: %v", err)
	}

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

func TestConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := newTestConfig(t, tmpDir, "concurrent-write.json")

	prd := &PRD{
		ProjectName: "Write Test",
		Version:     0,
	}

	if err := Save(cfg, prd); err != nil {
		t.Fatalf("initial Save failed: %v", err)
	}

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

	final, err := Load(cfg)
	if err != nil {
		t.Fatalf("final Load failed: %v", err)
	}

	if final.Version <= 1 {
		t.Errorf("expected version > 1, got %d", final.Version)
	}

	if final.ProjectName != "Modified" {
		t.Error("final PRD has corrupted data")
	}
}

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

func BenchmarkSave(b *testing.B) {
	tmpDir := b.TempDir()
	cfg := &config.Config{PRDFile: filepath.Join(tmpDir, "bench.json")}

	prd := &PRD{
		ProjectName: "Benchmark",
		Stories:     []*Story{{ID: "story-1", Title: "Test", Priority: 1, Slices: testSlice("works")}},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := Save(cfg, prd); err != nil {
			b.Fatalf("Save failed: %v", err)
		}
	}
}

func BenchmarkLoad(b *testing.B) {
	tmpDir := b.TempDir()
	cfg := &config.Config{PRDFile: filepath.Join(tmpDir, "bench.json")}

	prd := &PRD{
		ProjectName: "Benchmark",
		Stories:     []*Story{{ID: "story-1", Title: "Test", Priority: 1, Slices: testSlice("works")}},
	}

	Save(cfg, prd)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := Load(cfg); err != nil {
			b.Fatalf("Load failed: %v", err)
		}
	}
}
