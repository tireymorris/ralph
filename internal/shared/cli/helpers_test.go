package cli

import (
	"bytes"
	"strings"
	"testing"

	"ralph/internal/shared/prd"
)

func TestOutputPrefix(t *testing.T) {
	if got := OutputPrefix(false); got != "  " {
		t.Errorf("OutputPrefix(false) = %q, want %q", got, "  ")
	}
	if got := OutputPrefix(true); got != "  [!]" {
		t.Errorf("OutputPrefix(true) = %q, want %q", got, "  [!]")
	}
}

func TestStoryStatus(t *testing.T) {
	if got := StoryStatus(true); got != "[x]" {
		t.Errorf("StoryStatus(true) = %q, want %q", got, "[x]")
	}
	if got := StoryStatus(false); got != "[ ]" {
		t.Errorf("StoryStatus(false) = %q, want %q", got, "[ ]")
	}
}

func TestPrintStoryList(t *testing.T) {
	p := &prd.PRD{
		Stories: []*prd.Story{
			{Title: "Story 1", Priority: 1, Passes: true},
			{Title: "Story 2", Priority: 2, Passes: false},
		},
	}

	var buf bytes.Buffer
	PrintStoryList(&buf, p)

	output := buf.String()
	if !strings.Contains(output, "Story 1") {
		t.Error("output should contain story title")
	}
	if !strings.Contains(output, "[x]") {
		t.Error("output should show completed status")
	}
	if !strings.Contains(output, "[ ]") {
		t.Error("output should show incomplete status")
	}
	if !strings.Contains(output, "Stories:") {
		t.Error("output should contain 'Stories:' header")
	}
}

func TestPrintStoryDetails(t *testing.T) {
	p := &prd.PRD{
		Stories: []*prd.Story{
			{
				ID:          "story-1",
				Title:       "Test Story",
				Priority:    1,
				Description: "A test story",
				DependsOn:   []string{"story-0"},
				Slices: []*prd.Slice{
					{ID: "slice-1", Behavior: "AC1", RedHint: "red it"},
					{ID: "slice-2", Behavior: "AC2", RedHint: "red it again", RefactorHint: "extract helper"},
				},
			},
		},
	}

	var buf bytes.Buffer
	PrintStoryDetails(&buf, p)

	output := buf.String()
	if !strings.Contains(output, "Test Story") {
		t.Error("output should contain story title")
	}
	if !strings.Contains(output, "story-1") {
		t.Error("output should contain story ID")
	}
	if !strings.Contains(output, "A test story") {
		t.Error("output should contain description")
	}
	if !strings.Contains(output, "story-0") {
		t.Error("output should contain dependency")
	}
	if !strings.Contains(output, "Slices:") {
		t.Error("output should contain slice section")
	}
	if !strings.Contains(output, "extract helper") {
		t.Error("output should contain refactor hint")
	}
}

func TestPrintStoryDetailsNoDeps(t *testing.T) {
	p := &prd.PRD{
		Stories: []*prd.Story{
			{
				ID:       "story-1",
				Title:    "Simple Story",
				Priority: 1,
			},
		},
	}

	var buf bytes.Buffer
	PrintStoryDetails(&buf, p)

	output := buf.String()
	if strings.Contains(output, "Depends on") {
		t.Error("output should not contain 'Depends on' when no dependencies")
	}
}
