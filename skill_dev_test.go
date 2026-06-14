package main

import (
	"os"
	"strings"
	"testing"
)

func TestRalphSkillDocumentsCopilotRunner(t *testing.T) {
	data, err := os.ReadFile(".claude/skills/ralph/SKILL.md")
	if err != nil {
		t.Fatalf("read .claude/skills/ralph/SKILL.md: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "copilot") {
		t.Fatal(".claude/skills/ralph/SKILL.md must list copilot as a supported runner")
	}
	if !strings.Contains(content, "copilot login") {
		t.Fatal(".claude/skills/ralph/SKILL.md must document copilot login for auth")
	}
}
