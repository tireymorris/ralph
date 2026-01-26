package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"ralph/internal/config"
	"ralph/internal/prd"
	"ralph/internal/prompt"
)

func TestContextFieldRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{PRDFile: "prd.json", WorkDir: tmpDir}

	original := &prd.PRD{
		ProjectName: "Test Project",
		BranchName:  "feature/test",
		Context:     "Ruby 3.2 with RSpec. Tests in spec/ directory. Run with 'bundle exec rspec'. Main code in lib/.",
		TestSpec:    "Test end-to-end that the feature works correctly",
		Stories: []*prd.Story{
			{
				ID:                 "story-1",
				Title:              "Add feature",
				Description:        "Implement the feature",
				AcceptanceCriteria: []string{"Works correctly"},
				Priority:           1,
				Passes:             false,
				RetryCount:         0,
			},
		},
	}

	// Save the PRD
	if err := prd.Save(cfg, original); err != nil {
		t.Fatalf("Failed to save PRD: %v", err)
	}

	// Load it back
	loaded, err := prd.Load(cfg)
	if err != nil {
		t.Fatalf("Failed to load PRD: %v", err)
	}

	// Verify context was preserved
	if loaded.Context != original.Context {
		t.Errorf("Context not preserved.\nGot: %q\nWant: %q", loaded.Context, original.Context)
	}

	// Verify the JSON file contains the context field
	data, err := os.ReadFile(cfg.PRDPath())
	if err != nil {
		t.Fatalf("Failed to read PRD file: %v", err)
	}

	if !strings.Contains(string(data), `"context"`) {
		t.Error("JSON file should contain 'context' field")
	}
	if !strings.Contains(string(data), "Ruby 3.2 with RSpec") {
		t.Error("JSON file should contain the context value")
	}
}

func TestContextFieldOmittedWhenEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{PRDFile: "prd.json", WorkDir: tmpDir}

	original := &prd.PRD{
		ProjectName: "Test Project",
		BranchName:  "feature/test",
		Context:     "", // Empty context
		Stories: []*prd.Story{
			{ID: "story-1", Title: "Test", Priority: 1},
		},
	}

	if err := prd.Save(cfg, original); err != nil {
		t.Fatalf("Failed to save PRD: %v", err)
	}

	data, err := os.ReadFile(cfg.PRDPath())
	if err != nil {
		t.Fatalf("Failed to read PRD file: %v", err)
	}

	// With omitempty, empty context should not appear in JSON
	if strings.Contains(string(data), `"context"`) {
		t.Error("Empty context should be omitted from JSON (omitempty)")
	}
}

func TestStoryPromptIncludesContext(t *testing.T) {
	context := "Go 1.21 with standard testing. Tests alongside code as _test.go. Run with 'go test ./...'."

	result := prompt.StoryImplementation(
		"story-1",
		"Add feature",
		"Implement it",
		[]string{"It works"},
		"Test it",
		context,
		"prd.json",
		1, 0, 3,
	)

	if !strings.Contains(result, "CODEBASE CONTEXT") {
		t.Error("Prompt should contain 'CODEBASE CONTEXT' section")
	}
	if !strings.Contains(result, "Go 1.21 with standard testing") {
		t.Error("Prompt should include the context content")
	}
	if !strings.Contains(result, "go test ./...") {
		t.Error("Prompt should include test command from context")
	}
}

func TestStoryPromptOmitsContextSectionWhenEmpty(t *testing.T) {
	result := prompt.StoryImplementation(
		"story-1",
		"Add feature",
		"Implement it",
		[]string{"It works"},
		"Test it",
		"", // Empty context
		"prd.json",
		1, 0, 3,
	)

	if strings.Contains(result, "CODEBASE CONTEXT") {
		t.Error("Prompt should NOT contain 'CODEBASE CONTEXT' section when context is empty")
	}
}

func TestPRDGenerationPromptMentionsContext(t *testing.T) {
	result := prompt.PRDGeneration("Add auth", "prd.json", "feature")

	if !strings.Contains(result, `"context"`) {
		t.Error("PRD generation prompt should mention context field")
	}
	if !strings.Contains(result, "CONTEXT FIELD REQUIREMENTS") {
		t.Error("PRD generation prompt should include context field requirements")
	}
	if !strings.Contains(result, "Language/framework") {
		t.Error("PRD generation prompt should mention capturing language/framework")
	}
	if !strings.Contains(result, "Testing approach") {
		t.Error("PRD generation prompt should mention capturing testing approach")
	}
}

func TestBackwardsCompatibilityWithoutContext(t *testing.T) {
	// Test that PRDs without context field still load correctly
	tmpDir := t.TempDir()
	cfg := &config.Config{PRDFile: "prd.json", WorkDir: tmpDir}
	prdFile := cfg.PRDPath()

	// Write a PRD JSON without the context field (old format)
	oldFormatJSON := `{
		"project_name": "Old Project",
		"branch_name": "feature/old",
		"stories": [
			{
				"id": "story-1",
				"title": "Old Story",
				"description": "Desc",
				"acceptance_criteria": ["AC"],
				"priority": 1,
				"passes": false,
				"retry_count": 0
			}
		]
	}`

	if err := os.WriteFile(prdFile, []byte(oldFormatJSON), 0644); err != nil {
		t.Fatalf("Failed to write old format PRD: %v", err)
	}

	loaded, err := prd.Load(cfg)
	if err != nil {
		t.Fatalf("Failed to load old format PRD: %v", err)
	}

	if loaded.ProjectName != "Old Project" {
		t.Errorf("ProjectName = %q, want %q", loaded.ProjectName, "Old Project")
	}
	if loaded.Context != "" {
		t.Errorf("Context should be empty for old format PRD, got %q", loaded.Context)
	}
	if len(loaded.Stories) != 1 {
		t.Errorf("Should have 1 story, got %d", len(loaded.Stories))
	}
}

func TestContextPreservedThroughMultipleSaves(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{PRDFile: "prd.json", WorkDir: tmpDir}

	p := &prd.PRD{
		ProjectName: "Test",
		Context:     "Python 3.11 with pytest. Tests in tests/ directory.",
		Stories: []*prd.Story{
			{ID: "story-1", Title: "S1", Priority: 1, Passes: false},
			{ID: "story-2", Title: "S2", Priority: 2, Passes: false},
		},
	}

	// Save, load, modify, save again (simulating workflow)
	if err := prd.Save(cfg, p); err != nil {
		t.Fatalf("Save 1 failed: %v", err)
	}

	loaded, err := prd.Load(cfg)
	if err != nil {
		t.Fatalf("Load 1 failed: %v", err)
	}

	// Mark first story as complete
	loaded.Stories[0].Passes = true

	if err := prd.Save(cfg, loaded); err != nil {
		t.Fatalf("Save 2 failed: %v", err)
	}

	// Load again and verify context is still there
	final, err := prd.Load(cfg)
	if err != nil {
		t.Fatalf("Load 2 failed: %v", err)
	}

	if final.Context != p.Context {
		t.Errorf("Context lost after save/load cycle.\nGot: %q\nWant: %q", final.Context, p.Context)
	}
	if !final.Stories[0].Passes {
		t.Error("Story 1 should still be marked as passed")
	}
}

func TestContextJSONFormatting(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.Config{PRDFile: "prd.json", WorkDir: tmpDir}

	p := &prd.PRD{
		ProjectName: "Test",
		BranchName:  "feature/test",
		Context:     "Line 1.\nLine 2.\nLine 3.",
		Stories:     []*prd.Story{{ID: "s1", Title: "T", Priority: 1}},
	}

	if err := prd.Save(cfg, p); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify it's valid JSON
	data, _ := os.ReadFile(cfg.PRDPath())
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Saved PRD is not valid JSON: %v", err)
	}

	// Verify context with newlines is preserved
	loaded, _ := prd.Load(cfg)
	if loaded.Context != p.Context {
		t.Errorf("Context with newlines not preserved.\nGot: %q\nWant: %q", loaded.Context, p.Context)
	}
}
