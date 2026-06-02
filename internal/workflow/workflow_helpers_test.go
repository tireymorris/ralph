package workflow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkdirContainsSource(t *testing.T) {
	t.Run("empty directory", func(t *testing.T) {
		dir := t.TempDir()
		if workdirContainsSource(dir) {
			t.Error("expected empty directory to not report source")
		}
	})

	t.Run("directory with Go files", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
		if !workdirContainsSource(dir) {
			t.Error("expected directory with .go file to report source")
		}
	})

	t.Run("directory with only non-code files", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0644)
		os.WriteFile(filepath.Join(dir, "config.json"), []byte("{}"), 0644)
		if workdirContainsSource(dir) {
			t.Error("expected directory with only non-code files to not report source")
		}
	})

	t.Run("skips hidden directories", func(t *testing.T) {
		dir := t.TempDir()
		hiddenDir := filepath.Join(dir, ".git")
		os.MkdirAll(hiddenDir, 0755)
		os.WriteFile(filepath.Join(hiddenDir, "hooks.py"), []byte("# hook"), 0644)
		if workdirContainsSource(dir) {
			t.Error("expected code in hidden dirs to be ignored")
		}
	})

	t.Run("finds code in subdirectories", func(t *testing.T) {
		dir := t.TempDir()
		subDir := filepath.Join(dir, "src")
		os.MkdirAll(subDir, 0755)
		os.WriteFile(filepath.Join(subDir, "app.ts"), []byte("export {}"), 0644)
		if !workdirContainsSource(dir) {
			t.Error("expected code in subdirectory to report source")
		}
	})

	t.Run("empty string work dir", func(t *testing.T) {
		if workdirContainsSource("") {
			t.Error("expected empty work dir to not report source")
		}
	})
}
