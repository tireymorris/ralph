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

func readmeWorkflowSection(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	content := string(data)
	const heading = "## Workflow"
	start := strings.Index(content, heading)
	if start < 0 {
		t.Fatal("README.md must have a ## Workflow section")
	}
	rest := content[start+len(heading):]
	if next := strings.Index(rest, "\n## "); next >= 0 {
		rest = rest[:next]
	}
	return rest
}

func TestReadmeCleanupSectionDescribesFoldedReviewFlow(t *testing.T) {
	section := readmeWorkflowSection(t)
	for _, want := range []string{
		"PhaseCleanup",
		"waiting_implementation_review",
		"cleanup sub-state",
		"impl_review",
	} {
		if !strings.Contains(section, want) {
			t.Fatalf("README.md Workflow section must include %q", want)
		}
	}
	for _, forbidden := range []string{
		"implementation-review phase",
		"implementation review phase",
	} {
		if strings.Contains(strings.ToLower(section), forbidden) {
			t.Fatalf("README.md Workflow section must not describe a standalone %q", forbidden)
		}
	}

	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "run stays in PhaseCleanup") {
		t.Fatal("README.md Web API table must document cleanup review continue stays in PhaseCleanup")
	}
}

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
