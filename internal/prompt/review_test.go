package prompt

import (
	"strings"
	"testing"
)

func TestCriticalDiffReviewIncludesChangedFilesAndPRD(t *testing.T) {
	got := CriticalDiffReview("stack: Go", "prd.json", []string{"internal/foo.go", "web/bar.ts"})
	for _, want := range []string{
		"stack: Go",
		"prd.json",
		"internal/foo.go",
		"web/bar.ts",
		"CHANGED FILES:",
		KindDiffReview,
		"===ralph-findings===",
		"===/ralph-findings===",
		"category",
		"path",
		"summary",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("CriticalDiffReview() missing %q", want)
		}
	}
}
