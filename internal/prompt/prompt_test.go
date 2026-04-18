package prompt

import (
	"strings"
	"testing"
)

func TestClarifyingQuestions(t *testing.T) {
	tests := []struct {
		name            string
		userPrompt      string
		questionsFile   string
		isEmptyCodebase bool
		mustInclude     []string
		mustNotInclude  []string
	}{
		{
			name:            "existing codebase",
			userPrompt:      "Add user authentication",
			questionsFile:   ".ralph_questions.json",
			isEmptyCodebase: false,
			mustInclude: []string{
				"Add user authentication",
				".ralph_questions.json",
				"existing codebase",
				"2-5",
				"JSON file",
			},
			mustNotInclude: []string{"new project"},
		},
		{
			name:            "new project",
			userPrompt:      "Build a REST API",
			questionsFile:   ".ralph_questions.json",
			isEmptyCodebase: true,
			mustInclude: []string{
				"Build a REST API",
				"new project",
				".ralph_questions.json",
			},
			mustNotInclude: []string{"existing codebase"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClarifyingQuestions(tt.userPrompt, tt.questionsFile, tt.isEmptyCodebase)
			for _, phrase := range tt.mustInclude {
				if !strings.Contains(result, phrase) {
					t.Errorf("ClarifyingQuestions() missing %q", phrase)
				}
			}
			for _, phrase := range tt.mustNotInclude {
				if strings.Contains(result, phrase) {
					t.Errorf("ClarifyingQuestions() should not contain %q", phrase)
				}
			}
		})
	}
}

func TestPRDGenerationWithAnswers(t *testing.T) {
	t.Run("no answers produces same output as PRDGeneration", func(t *testing.T) {
		withNil := PRDGenerationWithAnswers("Add auth", "prd.json", "feature", false, nil)
		without := PRDGeneration("Add auth", "prd.json", "feature", false)
		if withNil != without {
			t.Error("PRDGenerationWithAnswers(nil) should equal PRDGeneration()")
		}
	})

	t.Run("answers are included in output", func(t *testing.T) {
		qas := []QuestionAnswer{
			{Question: "What language?", Answer: "Go"},
			{Question: "Need auth?", Answer: "JWT"},
		}
		result := PRDGenerationWithAnswers("Build API", "prd.json", "feature", false, qas)
		for _, phrase := range []string{
			"USER CLARIFICATIONS",
			"Q1: What language?",
			"A1: Go",
			"Q2: Need auth?",
			"A2: JWT",
		} {
			if !strings.Contains(result, phrase) {
				t.Errorf("PRDGenerationWithAnswers() missing %q", phrase)
			}
		}
	})

	t.Run("empty answers slice produces no clarifications section", func(t *testing.T) {
		result := PRDGenerationWithAnswers("Build API", "prd.json", "feature", false, []QuestionAnswer{})
		if strings.Contains(result, "USER CLARIFICATIONS") {
			t.Error("empty answers should not produce USER CLARIFICATIONS section")
		}
	})

	t.Run("answers appear before Write JSON instruction", func(t *testing.T) {
		qas := []QuestionAnswer{{Question: "Q?", Answer: "A"}}
		result := PRDGenerationWithAnswers("Test", "prd.json", "feature", false, qas)
		clarIdx := strings.Index(result, "USER CLARIFICATIONS")
		writeIdx := strings.Index(result, "Write JSON to")
		if clarIdx == -1 {
			t.Fatal("missing USER CLARIFICATIONS")
		}
		if clarIdx > writeIdx {
			t.Error("USER CLARIFICATIONS should appear before Write JSON instruction")
		}
	})
}

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
				nil, // dependsOn
				1,  // parallelCount
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
