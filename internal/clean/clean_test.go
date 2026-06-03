package clean

import (
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/workflow"
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

func TestRemoveState_removesPRDLock(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{WorkDir: dir, PRDFile: "prd.json"}

	lockPath := filepath.Join(dir, "prd.json.lock")
	if err := os.WriteFile(lockPath, nil, 0644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveState(cfg); err != nil {
		t.Fatalf("RemoveState: %v", err)
	}
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatalf("prd.json.lock still exists: %v", err)
	}
}

func TestRemoveState_removesClarifyingQuestions(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{WorkDir: dir, PRDFile: "prd.json"}

	questionsPath := filepath.Join(dir, workflow.ClarifyingQuestionsFile)
	if err := os.WriteFile(questionsPath, []byte("[]"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveState(cfg); err != nil {
		t.Fatalf("RemoveState: %v", err)
	}
	if _, err := os.Stat(questionsPath); !os.IsNotExist(err) {
		t.Fatalf("%s still exists: %v", workflow.ClarifyingQuestionsFile, err)
	}
}
