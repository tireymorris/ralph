package prompt

import (
	"strings"
	"testing"
)

func TestPRDGeneration(t *testing.T) {
	tests := []struct {
		name        string
		userPrompt  string
		mustInclude []string
	}{
		{
			name:       "basic prompt",
			userPrompt: "Add authentication",
			mustInclude: []string{
				"Add authentication",
				"project_name",
				"stories",
				"test_spec",
				"acceptance_criteria",
				"JSON",
			},
		},
		{
			name:       "complex prompt",
			userPrompt: "Build a REST API with user management and role-based access",
			mustInclude: []string{
				"Build a REST API with user management and role-based access",
				"priority",
			},
		},
		{
			name:       "instructs to analyze existing test conventions",
			userPrompt: "Add feature",
			mustInclude: []string{
				"Analyze existing test files to understand naming conventions",
				"Tests should follow existing project conventions",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PRDGeneration(tt.userPrompt)
			for _, phrase := range tt.mustInclude {
				if !strings.Contains(result, phrase) {
					t.Errorf("PRDGeneration() missing %q", phrase)
				}
			}
		})
	}
}

func TestStoryImplementation(t *testing.T) {
	tests := []struct {
		name               string
		title              string
		description        string
		acceptanceCriteria []string
		testSpec           string
		iteration          int
		completed          int
		total              int
		mustInclude        []string
	}{
		{
			name:               "basic story",
			title:              "Add login",
			description:        "Implement login functionality",
			acceptanceCriteria: []string{"User can login", "Error on bad credentials"},
			testSpec:           "Test login flow",
			iteration:          1,
			completed:          0,
			total:              3,
			mustInclude: []string{
				"Add login",
				"Implement login functionality",
				"User can login",
				"Error on bad credentials",
				"Test login flow",
				"Iteration 1",
				"0/3",
				"FOLLOW EXISTING TEST CONVENTIONS",
			},
		},
		{
			name:               "empty test spec uses default",
			title:              "Feature",
			description:        "Desc",
			acceptanceCriteria: []string{"AC"},
			testSpec:           "",
			iteration:          2,
			completed:          1,
			total:              2,
			mustInclude: []string{
				"No test spec provided",
				"create and run appropriate tests",
			},
		},
		{
			name:               "multiple acceptance criteria joined",
			title:              "T",
			description:        "D",
			acceptanceCriteria: []string{"A", "B", "C"},
			testSpec:           "spec",
			iteration:          1,
			completed:          0,
			total:              1,
			mustInclude:        []string{"A, B, C"},
		},
		{
			name:               "instructs to follow existing patterns",
			title:              "Add feature",
			description:        "Some feature",
			acceptanceCriteria: []string{"Works"},
			testSpec:           "Test it",
			iteration:          1,
			completed:          0,
			total:              1,
			mustInclude: []string{
				"FOLLOW EXISTING TEST CONVENTIONS",
				"DO NOT use generic names like",
				"story-1.test.js",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StoryImplementation(
				tt.title,
				tt.description,
				tt.acceptanceCriteria,
				tt.testSpec,
				tt.iteration,
				tt.completed,
				tt.total,
			)
			for _, phrase := range tt.mustInclude {
				if !strings.Contains(result, phrase) {
					t.Errorf("StoryImplementation() missing %q", phrase)
				}
			}
		})
	}
}
