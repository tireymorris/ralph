package workflow

import (
	"os"
	"path/filepath"
	"strings"
)

func isEmptyCodebase(workDir string) bool {
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
