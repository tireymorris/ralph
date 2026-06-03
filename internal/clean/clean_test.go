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

func TestRemoveState_removesOrphanedPRDTemps(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{WorkDir: dir, PRDFile: "prd.json"}

	tmpPath := filepath.Join(dir, ".prd.tmp.100.7")
	if err := os.WriteFile(tmpPath, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveState(cfg); err != nil {
		t.Fatalf("RemoveState: %v", err)
	}
	matches, err := filepath.Glob(filepath.Join(dir, ".prd.tmp.*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected no .prd.tmp.* files, got %v", matches)
	}
}

func TestRemoveState_removesRalphDir(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{WorkDir: dir, PRDFile: "prd.json"}

	metaPath := filepath.Join(dir, ".ralph", "runs", "x", "meta.json")
	if err := os.MkdirAll(filepath.Dir(metaPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(metaPath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveState(cfg); err != nil {
		t.Fatalf("RemoveState: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".ralph")); !os.IsNotExist(err) {
		t.Fatalf(".ralph still exists: %v", err)
	}
}

func seedAllStateArtifacts(t *testing.T, dir string) {
	t.Helper()
	paths := []string{
		filepath.Join(dir, "prd.json"),
		filepath.Join(dir, "prd.json.lock"),
		filepath.Join(dir, workflow.ClarifyingQuestionsFile),
		filepath.Join(dir, ".prd.tmp.100.7"),
		filepath.Join(dir, ".ralph", "runs", "x", "meta.json"),
	}
	for _, p := range paths {
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRemoveState_allArtifacts(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{WorkDir: dir, PRDFile: "prd.json"}
	seedAllStateArtifacts(t, dir)

	if err := RemoveState(cfg); err != nil {
		t.Fatalf("RemoveState: %v", err)
	}
	for _, name := range []string{"prd.json", "prd.json.lock", workflow.ClarifyingQuestionsFile, ".ralph"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("%s still exists: %v", name, err)
		}
	}
	matches, err := filepath.Glob(filepath.Join(dir, ".prd.tmp.*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected no .prd.tmp.* files, got %v", matches)
	}
}

func TestRemoveState_idempotent(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{WorkDir: dir, PRDFile: "prd.json"}

	seedAllStateArtifacts(t, dir)
	if err := RemoveState(cfg); err != nil {
		t.Fatalf("first RemoveState: %v", err)
	}
	if err := RemoveState(cfg); err != nil {
		t.Fatalf("second RemoveState: %v", err)
	}
}

func TestRemoveState_idempotentOnEmptyDir(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{WorkDir: dir, PRDFile: "prd.json"}

	for i := 1; i <= 2; i++ {
		if err := RemoveState(cfg); err != nil {
			t.Fatalf("RemoveState call %d: %v", i, err)
		}
	}
}
