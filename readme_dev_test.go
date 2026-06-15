package main

import (
	"os"
	"strings"
	"testing"
)

func readmeStateFilesSection(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	content := string(data)
	const heading = "## State files"
	start := strings.Index(content, heading)
	if start < 0 {
		t.Fatal("README.md must have a ## State files section")
	}
	rest := content[start+len(heading):]
	if next := strings.Index(rest, "\n## "); next >= 0 {
		rest = rest[:next]
	}
	return rest
}

func TestReadmeRunnersListsCopilot(t *testing.T) {
	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "copilot") {
		t.Fatal("README.md must list copilot as a supported runner")
	}
	if !strings.Contains(content, "copilot login") {
		t.Fatal("README.md must document copilot login for auth")
	}
}

func TestReadmeDocumentsBackups(t *testing.T) {
	section := readmeStateFilesSection(t)
	if !strings.Contains(section, "backups") {
		t.Fatal("README.md State files section must document backups")
	}
}
