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

func TestAgentsSupportedBackendsListsCopilot(t *testing.T) {
	data, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	start := strings.Index(string(data), "Supported backends:")
	if start < 0 {
		t.Fatal("AGENTS.md must have a Supported backends section")
	}
	rest := string(data)[start:]
	if next := strings.Index(rest, "\n## "); next >= 0 {
		rest = rest[:next]
	}
	if !strings.Contains(rest, "copilot") {
		t.Fatal("AGENTS.md Supported backends section must list copilot")
	}
}

func TestAgentsDocumentsBackups(t *testing.T) {
	section := agentsStateFilesSection(t)
	if !strings.Contains(section, "backups") {
		t.Fatal("AGENTS.md State files section must document backups")
	}
}
