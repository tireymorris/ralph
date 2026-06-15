package prompt

import (
	"embed"
	"io/fs"
	"strings"
	"testing"
)

func TestTemplatesParseAtInit(t *testing.T) {
	if templates == nil {
		t.Fatal("templates should be parsed at init")
	}
	names := []string{
		"clarify",
		"prd-generate",
		"prd-self-review",
		"prd-critique-revision",
		"prd-clarification-revision",
		"story-implement",
		"diff-review",
		"recovery",
		"cleanup",
		"followup",
		"codebase-context",
		"changed-files",
		"clarifications",
		"commit-rules",
		"working-conventions",
		"review-conventions",
		"refactor-discipline",
	}
	for _, name := range names {
		if templates.Lookup(name) == nil {
			t.Errorf("template %q not found", name)
		}
	}
}

func TestAllTemplateFilesEmbedded(t *testing.T) {
	count := 0
	err := fs.WalkDir(templateFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".tmpl") {
			count++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk templates: %v", err)
	}
	if count < 17 {
		t.Fatalf("expected at least 17 template files, got %d", count)
	}
}

func TestRecoveryTemplateContainsMarker(t *testing.T) {
	got := RecoverFromFailure("", "prd.json", RecoveryReasonStoryFailure, 1, 2, "boom", nil, nil)
	if !strings.Contains(got, RecoveryAgentMarker) {
		t.Fatalf("rendered recovery prompt must contain RecoveryAgentMarker %q", RecoveryAgentMarker)
	}
}

func TestAllTemplatesExecuteWithMinimalData(t *testing.T) {
	cases := []struct {
		name string
		fn   func() string
	}{
		{"clarify", func() string {
			return ClarifyingQuestions("build x", ".ralph/questions.json", false)
		}},
		{"prd-generate", func() string {
			return PRDGeneration("build x", "prd.json", "feature", false)
		}},
		{"prd-self-review", func() string {
			return PRDSelfReview("build x", "prd.json", 1, 3)
		}},
		{"prd-critique-revision", func() string {
			return PRDCritiqueRevision("build x", "prd.json", "more tests")
		}},
		{"prd-clarification-revision", func() string {
			return PRDClarificationRevision("build x", "prd.json", []QuestionAnswer{{Question: "Q?", Answer: "A"}})
		}},
		{"story-implement", func() string {
			return StoryImplementation("story-1", "Title", "Desc", []SliceData{{ID: "slice-1", Behavior: "done", RedHint: "red"}}, "", "", "prd.json", 0, 1, nil)
		}},
		{"diff-review", func() string {
			return CriticalDiffReview("", "prd.json", nil)
		}},
		{"recovery", func() string {
			return RecoverFromFailure("", "prd.json", RecoveryReasonStoryFailure, 1, 2, "boom", nil, nil)
		}},
		{"cleanup", func() string {
			return Cleanup("", "prd.json", nil)
		}},
		{"followup", func() string {
			return FollowUpRevision("add tests", "prd.json", "event log")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.fn(); got == "" {
				t.Fatal("expected non-empty prompt")
			}
		})
	}
}

// Ensure embed package is referenced for TestAllTemplateFilesEmbedded.
var _ embed.FS = templateFS
