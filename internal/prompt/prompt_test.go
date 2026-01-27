package prompt

import (
	"strings"
	"testing"
)

func TestPRDGeneration(t *testing.T) {
	tests := []struct {
		name            string
		userPrompt      string
		prdFile         string
		branchPrefix    string
		isEmptyCodebase bool
		mustInclude     []string
		mustNotInclude  []string
	}{
		{
			name:            "basic prompt with existing codebase",
			userPrompt:      "Add authentication",
			prdFile:         "prd.json",
			branchPrefix:    "feature",
			isEmptyCodebase: false,
			mustInclude: []string{
				"Add authentication",
				"project_name",
				"context",
				"stories",
				"prd.json",
				"feature/",
				`"version": 1`,
				"ACTUALLY observe",
			},
			mustNotInclude: []string{
				"no existing source code",
			},
		},
		{
			name:            "empty codebase prompt",
			userPrompt:      "Build a REST API",
			prdFile:         "prd.json",
			branchPrefix:    "feature",
			isEmptyCodebase: true,
			mustInclude: []string{
				"Build a REST API",
				`"version": 1`,
				"no existing source code",
				"Do NOT assume or invent",
			},
			mustNotInclude: []string{
				"ACTUALLY observe",
			},
		},
		{
			name:            "custom prd file",
			userPrompt:      "Add feature",
			prdFile:         "custom.json",
			branchPrefix:    "feat",
			isEmptyCodebase: false,
			mustInclude: []string{
				"custom.json",
				"feat/",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PRDGeneration(tt.userPrompt, tt.prdFile, tt.branchPrefix, tt.isEmptyCodebase)
			for _, phrase := range tt.mustInclude {
				if !strings.Contains(result, phrase) {
					t.Errorf("PRDGeneration() missing %q", phrase)
				}
			}
			for _, phrase := range tt.mustNotInclude {
				if strings.Contains(result, phrase) {
					t.Errorf("PRDGeneration() should not contain %q", phrase)
				}
			}
		})
	}
}

func TestPRDValidation(t *testing.T) {
	tests := []struct {
		name        string
		prdJSON     string
		prdFile     string
		context     string
		mustInclude []string
	}{
		{
			name:    "includes file path and context",
			prdJSON: `{"project_name":"Test"}`,
			prdFile: "prd.json",
			context: "Go 1.21 with standard testing",
			mustInclude: []string{
				"prd.json",
				"Go 1.21 with standard testing",
				"CODEBASE CONTEXT",
				`{"project_name":"Test"}`,
			},
		},
		{
			name:    "omits context section when empty",
			prdJSON: `{"project_name":"Test"}`,
			prdFile: "custom.json",
			context: "",
			mustInclude: []string{
				"custom.json",
				`{"project_name":"Test"}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PRDValidation(tt.prdJSON, tt.prdFile, tt.context)
			for _, phrase := range tt.mustInclude {
				if !strings.Contains(result, phrase) {
					t.Errorf("PRDValidation() missing %q", phrase)
				}
			}
			if tt.context == "" && strings.Contains(result, "CODEBASE CONTEXT") {
				t.Error("PRDValidation() should not contain CODEBASE CONTEXT when context is empty")
			}
		})
	}
}

func TestStoryImplementation(t *testing.T) {
	tests := []struct {
		name               string
		storyID            string
		title              string
		description        string
		acceptanceCriteria []string
		featureTestSpec    string
		context            string
		prdFile            string
		iteration          int
		completed          int
		total              int
		mustInclude        []string
		mustNotInclude     []string
	}{
		{
			name:               "basic story without context or test spec",
			storyID:            "story-1",
			title:              "Add login",
			description:        "Implement login functionality",
			acceptanceCriteria: []string{"User can login", "Error on bad credentials"},
			featureTestSpec:    "",
			context:            "",
			prdFile:            "prd.json",
			iteration:          1,
			completed:          0,
			total:              3,
			mustInclude: []string{
				"Add login",
				"story-1",
				"Implement login functionality",
				"User can login",
				"Error on bad credentials",
				"Iteration 1",
				"0/3",
				"prd.json",
			},
			mustNotInclude: []string{"CODEBASE CONTEXT", "FEATURE TEST SPEC"},
		},
		{
			name:               "story with context and feature test spec",
			storyID:            "story-1",
			title:              "Add feature",
			description:        "Implement feature",
			acceptanceCriteria: []string{"Works"},
			featureTestSpec:    "Test end-to-end: 1) Login works, 2) Errors handled",
			context:            "Ruby 3.2 with RSpec. Tests in spec/ directory. Run with 'bundle exec rspec'.",
			prdFile:            "prd.json",
			iteration:          1,
			completed:          0,
			total:              2,
			mustInclude: []string{
				"Add feature",
				"Implement feature",
				"Works",
				"CODEBASE CONTEXT",
				"Ruby 3.2 with RSpec",
				"bundle exec rspec",
				"FEATURE TEST SPEC",
				"Test end-to-end",
			},
		},
		{
			name:               "multiple acceptance criteria joined",
			storyID:            "story-3",
			title:              "T",
			description:        "D",
			acceptanceCriteria: []string{"A", "B", "C"},
			featureTestSpec:    "",
			context:            "",
			prdFile:            "prd.json",
			iteration:          1,
			completed:          0,
			total:              1,
			mustInclude:        []string{"A; B; C"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StoryImplementation(
				tt.storyID,
				tt.title,
				tt.description,
				tt.acceptanceCriteria,
				tt.featureTestSpec,
				tt.context,
				tt.prdFile,
				tt.iteration,
				tt.completed,
				tt.total,
			)
			for _, phrase := range tt.mustInclude {
				if !strings.Contains(result, phrase) {
					t.Errorf("StoryImplementation() missing %q in:\n%s", phrase, result)
				}
			}
			for _, phrase := range tt.mustNotInclude {
				if strings.Contains(result, phrase) {
					t.Errorf("StoryImplementation() should not contain %q", phrase)
				}
			}
		})
	}
}
