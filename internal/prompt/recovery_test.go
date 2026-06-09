package prompt

import (
	"strings"
	"testing"
)

func TestRecoverFromFailureIncludesFindingsAndReason(t *testing.T) {
	p := RecoverFromFailure(
		"Vue routes via js_routes.js.erb",
		"prd.json",
		RecoveryReasonReviewFindings,
		1,
		2,
		"",
		[]RecoveryFinding{{
			Category: "wrong-target",
			Path:     "app/javascript/routes.js",
			Summary:  "not imported anywhere",
		}},
		[]string{"app/javascript/routes.js"},
	)

	if !strings.Contains(p, RecoveryAgentMarker) {
		t.Fatal("expected recovery agent marker")
	}
	if !strings.Contains(p, "review_findings") {
		t.Fatal("expected recovery reason")
	}
	if !strings.Contains(p, "routes.js") {
		t.Fatal("expected finding path in prompt")
	}
	if !strings.Contains(p, "js_routes.js.erb") {
		t.Fatal("expected codebase context in prompt")
	}
}

func TestIsRecoveryPrompt(t *testing.T) {
	if !IsRecoveryPrompt(RecoverFromFailure("", "prd.json", RecoveryReasonStoryFailure, 1, 2, "boom", nil, nil)) {
		t.Fatal("expected recovery prompt detection")
	}
	if IsRecoveryPrompt("implement story foo") {
		t.Fatal("did not expect recovery prompt detection")
	}
}
