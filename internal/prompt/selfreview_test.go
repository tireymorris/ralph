package prompt

import (
	"strings"
	"testing"
)

func TestPRDSelfReview(t *testing.T) {
	result := PRDSelfReview("Add user authentication", "prd.json", 2, 3)

	mustInclude := []string{
		"Add user authentication",
		"prd.json",
		"round 2 of 3",
		".ralph_prd_review.json",
		`"approved"`,
		`"summary"`,
		"objectively verifiable",
		"must exist in the repo",
		"focused, additive diff",
		"failing tests first (TDD)",
		"depends_on",
		"context",
		"fewest touched lines",
	}
	for _, phrase := range mustInclude {
		if !strings.Contains(result, phrase) {
			t.Errorf("PRDSelfReview() missing %q", phrase)
		}
	}
}
