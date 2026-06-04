package main

import (
	"os"
	"strings"
	"testing"
)

func agentsStateFilesSection(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	content := string(data)
	const heading = "## State files"
	start := strings.Index(content, heading)
	if start < 0 {
		t.Fatal("AGENTS.md must have a ## State files section")
	}
	rest := content[start+len(heading):]
	if next := strings.Index(rest, "\n## "); next >= 0 {
		rest = rest[:next]
	}
	return rest
}

func TestAgentsDocumentsBackups(t *testing.T) {
	section := agentsStateFilesSection(t)
	if !strings.Contains(section, "backups") {
		t.Fatal("AGENTS.md State files section must document backups")
	}
}
