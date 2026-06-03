package clean

import (
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/shared/config"
)

func TestRemoveState_removesPRD(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{WorkDir: dir, PRDFile: "prd.json"}

	prdPath := filepath.Join(dir, "prd.json")
	if err := os.WriteFile(prdPath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveState(cfg); err != nil {
		t.Fatalf("RemoveState: %v", err)
	}
	if _, err := os.Stat(prdPath); !os.IsNotExist(err) {
		t.Fatalf("prd.json still exists: %v", err)
	}
}
