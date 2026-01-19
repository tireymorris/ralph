package prd

import (
	"os"
	"path/filepath"
	"testing"

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
