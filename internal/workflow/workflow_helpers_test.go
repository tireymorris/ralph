package workflow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsEmptyCodebase(t *testing.T) {
	t.Run("empty directory", func(t *testing.T) {
		dir := t.TempDir()
		if !isEmptyCodebase(dir) {
			t.Error("expected empty directory to be detected as empty codebase")
		}
	})

	t.Run("directory with Go files", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
		if isEmptyCodebase(dir) {
			t.Error("expected directory with .go file to be detected as non-empty")
		}
	})

	t.Run("directory with only non-code files", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0644)
		os.WriteFile(filepath.Join(dir, "config.json"), []byte("{}"), 0644)
		if !isEmptyCodebase(dir) {
			t.Error("expected directory with only non-code files to be detected as empty codebase")
		}
	})

	t.Run("skips hidden directories", func(t *testing.T) {
		dir := t.TempDir()
		hiddenDir := filepath.Join(dir, ".git")
		os.MkdirAll(hiddenDir, 0755)
		os.WriteFile(filepath.Join(hiddenDir, "hooks.py"), []byte("# hook"), 0644)
		if !isEmptyCodebase(dir) {
			t.Error("expected code in hidden dirs to be ignored")
		}
	})

	t.Run("finds code in subdirectories", func(t *testing.T) {
		dir := t.TempDir()
		subDir := filepath.Join(dir, "src")
		os.MkdirAll(subDir, 0755)
		os.WriteFile(filepath.Join(subDir, "app.ts"), []byte("export {}"), 0644)
		if isEmptyCodebase(dir) {
			t.Error("expected code in subdirectory to be detected")
		}
	})

	t.Run("empty string work dir", func(t *testing.T) {
		if !isEmptyCodebase("") {
			t.Error("expected empty work dir to return true")
		}
	})
}
