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
				"0-5",
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
			completed:          0,
			total:              3,
			mustInclude: []string{
				"Add login",
				"story-1",
				"Implement login functionality",
				"User can login",
				"Error on bad credentials",
				"0/3",
				"prd.json",
			},
			mustNotInclude: []string{"CODEBASE CONTEXT", "FEATURE TEST SPEC", "CRITIQUE"},
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
				tt.completed,
				tt.total,
				nil,
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

func TestPRDCritiqueRevisionIncludesCritique(t *testing.T) {
	result := PRDCritiqueRevision("add login", "prd.json", "Needs more tests")

	if !strings.Contains(result, "Needs more tests") {
		t.Fatal("PRDCritiqueRevision() should include critique text")
	}
	if !strings.Contains(result, "add login") {
		t.Fatal("PRDCritiqueRevision() should include user prompt")
	}
	if !strings.Contains(result, "prd.json") {
		t.Fatal("PRDCritiqueRevision() should reference PRD file")
	}
}

func TestCleanup_skip_wording(t *testing.T) {
	result := Cleanup("", "prd.json", nil)
	if !strings.Contains(result, "without modifying") || !strings.Contains(result, "without committing") {
		t.Errorf("Cleanup() should include no-changes-needed skip wording, got:\n%s", result)
	}
}

func TestCleanup_includes_improvement_guidance(t *testing.T) {
	result := Cleanup("Go 1.24 app", "prd.json", nil)
	if result == "" {
		t.Fatal("Cleanup() returned empty string")
	}
	for _, want := range []string{"SOLID", "DRY"} {
		if !strings.Contains(result, want) {
			t.Errorf("Cleanup() missing %q", want)
		}
	}
	hasCodebaseConventions := strings.Contains(result, "codebase conventions") || strings.Contains(result, "existing conventions")
	if !hasCodebaseConventions {
		t.Errorf("Cleanup() missing 'codebase conventions' or 'existing conventions'")
	}
	hasConsolidate := strings.Contains(result, "consolidate") || strings.Contains(result, "combine")
	if !hasConsolidate {
		t.Errorf("Cleanup() missing 'consolidate' or 'combine' referencing specs")
	}
}

func TestCleanup_instructs_running_tests_before_committing(t *testing.T) {
	result := Cleanup("", "prd.json", nil)
	hasTestInstruction := strings.Contains(result, "run") && strings.Contains(result, "test")
	if !hasTestInstruction {
		t.Errorf("Cleanup() should instruct running tests")
	}
	hasCommitInstruction := strings.Contains(result, "commit")
	if !hasCommitInstruction {
		t.Errorf("Cleanup() should reference committing")
	}
}

func TestCleanup_omits_context_section_when_empty(t *testing.T) {
	result := Cleanup("", "prd.json", nil)
	if strings.Contains(result, "CODEBASE CONTEXT") {
		t.Error("Cleanup() with empty context should not include CODEBASE CONTEXT section")
	}
}

func TestCleanup_includes_codebaseContext_when_nonempty(t *testing.T) {
	result := Cleanup("Go 1.24 with Bubble Tea", "prd.json", nil)
	if !strings.Contains(result, "Go 1.24 with Bubble Tea") {
		t.Errorf("Cleanup() should include codebaseContext value")
	}
}

func TestCleanup_includes_changed_files_when_provided(t *testing.T) {
	files := []string{"internal/foo/bar.go", "internal/foo/bar_test.go"}
	result := Cleanup("", "prd.json", files)
	if !strings.Contains(result, "CHANGED FILES") {
		t.Error("Cleanup() with changed files should include CHANGED FILES section")
	}
	for _, f := range files {
		if !strings.Contains(result, f) {
			t.Errorf("Cleanup() should include file %q", f)
		}
	}
	if !strings.Contains(result, "Only modify") {
		t.Error("Cleanup() should instruct limiting scope to changed files")
	}
}

func TestCleanup_omits_changed_files_section_when_empty(t *testing.T) {
	result := Cleanup("", "prd.json", nil)
	if strings.Contains(result, "CHANGED FILES") {
		t.Error("Cleanup() with nil changed files should not include CHANGED FILES section")
	}
	result2 := Cleanup("", "prd.json", []string{})
	if strings.Contains(result2, "CHANGED FILES") {
		t.Error("Cleanup() with empty changed files should not include CHANGED FILES section")
	}
}

func TestPRDClarificationRevisionIncludesAnswers(t *testing.T) {
	result := PRDClarificationRevision("add login", "prd.json", []QuestionAnswer{
		{Question: "Which auth?", Answer: "OAuth"},
	})

	if !strings.Contains(result, "Which auth?") || !strings.Contains(result, "OAuth") {
		t.Fatal("PRDClarificationRevision() should include clarifications")
	}
}
