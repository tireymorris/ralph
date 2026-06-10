package main

import (
	"os"
	"strings"
	"testing"
)

func TestGitignoreIgnoresRalphStatePaths(t *testing.T) {
	data, err := os.ReadFile(".gitignore")
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	content := string(data)
	for _, entry := range []string{"prd.json", "prd.json.lock", ".ralph/"} {
		if !strings.Contains(content, entry) {
			t.Fatalf(".gitignore must include Ralph state entry %q", entry)
		}
	}
}
