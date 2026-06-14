package main

import (
	"os"
	"strings"
	"testing"
)

func readmeUsageSection(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	content := string(data)
	const heading = "## Usage"
	start := strings.Index(content, heading)
	if start < 0 {
		t.Fatal("README.md must have a ## Usage section")
	}
	rest := content[start+len(heading):]
	if next := strings.Index(rest, "\n## "); next >= 0 {
		rest = rest[:next]
	}
	return rest
}


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

func TestReadmeDocumentsClean(t *testing.T) {
	usage := readmeUsageSection(t)
	if !strings.Contains(usage, "ralph clean") {
		t.Fatal("README Usage section must contain the exact command string ralph clean")
	}
	for _, want := range []string{"prd.json", ".ralph/prd.tmp.*", ".ralph/"} {
		if !strings.Contains(usage, want) {
			t.Fatalf("README Usage section must mention %q near ralph clean", want)
		}
	}

	section := readmeStateFilesSection(t)
	if !strings.Contains(section, ".ralph/prd.tmp.*") {
		t.Fatal("README State files section must document .ralph/prd.tmp.*")
	}
	if !strings.Contains(section, "ralph clean") {
		t.Fatal("README State files section must note that ralph clean removes artifacts idempotently")
	}
}

func TestReadmeDocumentsBackupOnNewRun(t *testing.T) {
	section := readmeStateFilesSection(t)
	if !strings.Contains(section, ".ralph/backups/") {
		t.Fatal("README State files section must document .ralph/backups/")
	}
}

func TestReadmeResumeDoesNotArchive(t *testing.T) {
	section := readmeStateFilesSection(t)
	if !strings.Contains(section, "--resume") {
		t.Fatal("README State files section must mention --resume")
	}
	lower := strings.ToLower(section)
	if !strings.Contains(lower, "do not archive") && !strings.Contains(lower, "does not archive") {
		t.Fatal("README State files section must state that --resume does not archive prior state")
	}
}

func TestReadmeCleanDeletesStateInPlace(t *testing.T) {
	section := readmeStateFilesSection(t)
	if !strings.Contains(section, "ralph clean") {
		t.Fatal("README State files section must document ralph clean")
	}
	if !strings.Contains(strings.ToLower(section), "delete") {
		t.Fatal("README State files section must describe ralph clean as deleting state")
	}
}

func TestReadmeRequiresListsCopilotRunner(t *testing.T) {
	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "Requires:") && strings.Contains(line, "copilot") {
			return
		}
	}
	t.Fatal("README.md Requires line must list copilot as a supported runner")
}

func TestReadmeDocumentsReleaseBuild(t *testing.T) {
	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "scripts/build.sh") {
		t.Fatal("README.md must document scripts/build.sh for release-style builds")
	}
}
