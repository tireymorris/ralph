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
		storyID            string
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
			storyID:            "story-1",
			iteration:          1,
			completed:          0,
			total:              3,
			mustInclude: []string{
				"Add login",
				"Implement login functionality",
				"User can login",
				"Error on bad credentials",
				"Test login flow",
				"story-1",
				"Iteration 1",
				"0/3",
			},
		},
		{
			name:               "empty test spec uses default",
			title:              "Feature",
			description:        "Desc",
			acceptanceCriteria: []string{"AC"},
			testSpec:           "",
			storyID:            "s1",
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
			storyID:            "s",
			iteration:          1,
			completed:          0,
			total:              1,
			mustInclude:        []string{"A, B, C"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StoryImplementation(
				tt.title,
				tt.description,
				tt.acceptanceCriteria,
				tt.testSpec,
				tt.storyID,
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
