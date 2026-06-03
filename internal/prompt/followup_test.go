package prompt

import (
	"strings"
	"testing"
)

func TestFollowUpRevisionIncludesInputs(t *testing.T) {
	result := FollowUpRevision("add dark mode", "prd.json", "event: output line")

	for _, phrase := range []string{
		"add dark mode",
		"prd.json",
		"event: output line",
		"Run transcript",
	} {
		if !strings.Contains(result, phrase) {
			t.Errorf("FollowUpRevision() missing %q", phrase)
		}
	}
}
