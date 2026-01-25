package prompt

import (
	"strings"
	"testing"
)

func TestPRDGeneration(t *testing.T) {
	tests := []struct {
		name         string
		userPrompt   string
		prdFile      string
		branchPrefix string
		mustInclude  []string
	}{
		{
			name:         "basic prompt",
			userPrompt:   "Add authentication",
			prdFile:      "prd.json",
			branchPrefix: "feature",
			mustInclude: []string{
				"Add authentication",
				"project_name",
				"context",
				"stories",
				"prd.json",
				"feature/",
				"CONTEXT FIELD REQUIREMENTS",
			},
		},
		{
			name:         "complex prompt",
			userPrompt:   "Build a REST API with user management and role-based access",
			prdFile:      "prd.json",
			branchPrefix: "feature",
			mustInclude: []string{
				"Build a REST API with user management and role-based access",
				"priority",
			},
		},
		{
			name:         "custom prd file",
			userPrompt:   "Add feature",
			prdFile:      "custom.json",
			branchPrefix: "feat",
			mustInclude: []string{
				"custom.json",
				"feat/",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PRDGeneration(tt.userPrompt, tt.prdFile, tt.branchPrefix)
			for _, phrase := range tt.mustInclude {
				if !strings.Contains(result, phrase) {
					t.Errorf("PRDGeneration() missing %q", phrase)
				}
			}
		})
	}
}

func TestJSONRepair(t *testing.T) {
	tests := []struct {
		name        string
		prdFile     string
		parseError  string
		mustInclude []string
	}{
		{
			name:       "basic repair prompt",
			prdFile:    "prd.json",
			parseError: "invalid character ']' after object key:value pair",
			mustInclude: []string{
				"prd.json",
				"invalid character ']'",
				"fix the JSON syntax error",
				"Missing or extra commas",
			},
		},
		{
			name:       "custom file",
			prdFile:    "custom.json",
			parseError: "unexpected end of JSON input",
			mustInclude: []string{
				"custom.json",
				"unexpected end of JSON",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JSONRepair(tt.prdFile, tt.parseError)
			for _, phrase := range tt.mustInclude {
				if !strings.Contains(result, phrase) {
					t.Errorf("JSONRepair() missing %q", phrase)
				}
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
		testSpec           string
		context            string
		prdFile            string
		iteration          int
		completed          int
		total              int
		mustInclude        []string
		mustNotInclude     []string
	}{
		{
			name:               "basic story without context",
			storyID:            "story-1",
			title:              "Add login",
			description:        "Implement login functionality",
			acceptanceCriteria: []string{"User can login", "Error on bad credentials"},
			testSpec:           "Test login flow",
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
				"Test login flow",
				"Iteration 1",
				"0/3",
				"prd.json",
			},
			mustNotInclude: []string{"CODEBASE CONTEXT"},
		},
		{
			name:               "story with context",
			storyID:            "story-1",
			title:              "Add feature",
			description:        "Implement feature",
			acceptanceCriteria: []string{"Works"},
			testSpec:           "Test it",
			context:            "Ruby 3.2 with RSpec. Tests in spec/ directory. Run with 'bundle exec rspec'.",
			prdFile:            "prd.json",
			iteration:          1,
			completed:          0,
			total:              2,
			mustInclude: []string{
				"CODEBASE CONTEXT",
				"Ruby 3.2 with RSpec",
				"Tests in spec/ directory",
				"bundle exec rspec",
			},
		},
		{
			name:               "empty test spec uses default",
			storyID:            "story-2",
			title:              "Feature",
			description:        "Desc",
			acceptanceCriteria: []string{"AC"},
			testSpec:           "",
			context:            "",
			prdFile:            "prd.json",
			iteration:          2,
			completed:          1,
			total:              2,
			mustInclude: []string{
				"Create and run appropriate tests",
			},
		},
		{
			name:               "multiple acceptance criteria joined",
			storyID:            "story-3",
			title:              "T",
			description:        "D",
			acceptanceCriteria: []string{"A", "B", "C"},
			testSpec:           "spec",
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
				tt.testSpec,
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
