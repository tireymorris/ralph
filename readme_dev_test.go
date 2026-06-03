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

func TestReadmeUsageDocumentsClean(t *testing.T) {
	usage := readmeUsageSection(t)
	if !strings.Contains(usage, "ralph clean") {
		t.Fatal("README Usage section must contain the exact command string ralph clean")
	}
}

func TestReadmeUsageDocumentsCleanRemovesState(t *testing.T) {
	usage := readmeUsageSection(t)
	for _, want := range []string{"prd.json", ".prd.tmp.*", ".ralph/"} {
		if !strings.Contains(usage, want) {
			t.Fatalf("README Usage section must mention %q near ralph clean", want)
		}
	}
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
