package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// legacyIsEmptyCodebase mirrors main-branch isEmptyCodebase before the workdirContainsSource rename.
func legacyIsEmptyCodebase(workDir string) bool {
	if workDir == "" {
		return true
	}

	sourceExts := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
		".rb": true, ".java": true, ".rs": true, ".c": true, ".cpp": true, ".cs": true,
		".php": true, ".swift": true, ".kt": true, ".ex": true, ".hs": true, ".scala": true,
		".sh": true, ".ml": true, ".r": true, ".pl": true, ".lua": true, ".dart": true,
		".vue": true, ".svelte": true, ".html": true, ".css": true, ".scss": true,
	}

	found := false
	filepath.WalkDir(workDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if sourceExts[ext] {
			found = true
			return filepath.SkipAll
		}
		return nil
	})

	return !found
}

func TestWorkdirContainsSourceMatchesLegacyIsEmpty(t *testing.T) {
	emptyDir := t.TempDir()
	withGo := t.TempDir()
	os.WriteFile(filepath.Join(withGo, "main.go"), []byte("package main"), 0644)

	onlyDocs := t.TempDir()
	os.WriteFile(filepath.Join(onlyDocs, "README.md"), []byte("# test"), 0644)

	withHiddenCode := t.TempDir()
	hiddenDir := filepath.Join(withHiddenCode, ".git")
	os.MkdirAll(hiddenDir, 0755)
	os.WriteFile(filepath.Join(hiddenDir, "hooks.py"), []byte("# hook"), 0644)

	withNestedCode := t.TempDir()
	subDir := filepath.Join(withNestedCode, "src")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "app.ts"), []byte("export {}"), 0644)

	cases := []struct {
		name   string
		workDir string
	}{
		{name: "empty directory", workDir: emptyDir},
		{name: "directory with Go files", workDir: withGo},
		{name: "directory with only non-code files", workDir: onlyDocs},
		{name: "skips hidden directories", workDir: withHiddenCode},
		{name: "finds code in subdirectories", workDir: withNestedCode},
		{name: "empty string work dir", workDir: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			legacy := legacyIsEmptyCodebase(tc.workDir)
			got := !workdirContainsSource(tc.workDir)
			if got != legacy {
				t.Fatalf("!workdirContainsSource(%q) = %v, legacy isEmptyCodebase = %v", tc.workDir, got, legacy)
			}
		})
	}
}

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
