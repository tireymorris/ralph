package prompt

import "testing"

func TestRenderedPromptsIncludeKindMarker(t *testing.T) {
	cases := []struct {
		name string
		got  string
		kind string
	}{
		{"clarify", ClarifyingQuestions("build x", ".ralph/questions.json", false), KindClarify},
		{"prd-generate", PRDGeneration("build x", "prd.json", "feature", false), KindPRDGenerate},
		{"story-implement", StoryImplementation("story-1", "Title", "Desc", []SliceData{{ID: "slice-1", Behavior: "b", RedHint: "r"}}, "", "", "prd.json", 0, 1, nil), KindStoryImplement},
		{"diff-review", CriticalDiffReview("", "prd.json", nil), KindDiffReview},
		{"recovery", RecoverFromFailure("", "prd.json", RecoveryReasonStoryFailure, 1, 2, "boom", nil, nil), KindRecovery},
		{"cleanup", Cleanup("", "prd.json", nil), KindCleanup},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got == "" {
				t.Fatal("expected non-empty prompt")
			}
			if !HasKind(tc.got, tc.kind) {
				t.Fatalf("prompt missing kind %q:\n%s", tc.kind, tc.got)
			}
			if got := Kind(tc.got); got != tc.kind {
				t.Fatalf("Kind() = %q, want %q", got, tc.kind)
			}
		})
	}
}

func TestHasKindRejectsUnmarkedPrompts(t *testing.T) {
	if HasKind("implement story foo", KindStoryImplement) {
		t.Fatal("unmarked prompt should not match kind")
	}
	if Kind("plain text") != "" {
		t.Fatalf("Kind() = %q, want empty for unmarked prompt", Kind("plain text"))
	}
}

func TestIsRecoveryPromptUsesKindMarker(t *testing.T) {
	p := RecoverFromFailure("", "prd.json", RecoveryReasonStoryFailure, 1, 2, "boom", nil, nil)
	if !IsRecoveryPrompt(p) {
		t.Fatal("IsRecoveryPrompt() = false, want true for recovery prompt")
	}
	if IsRecoveryPrompt("Ralph's recovery agent without marker") {
		t.Fatal("IsRecoveryPrompt() = true for prose-only match, want false")
	}
}
