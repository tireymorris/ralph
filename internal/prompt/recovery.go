package prompt

import (
	"encoding/json"
	"fmt"
	"strings"
)

const RecoveryAgentMarker = "Ralph's recovery agent"

type RecoveryReason string

const (
	RecoveryReasonReviewFindings    RecoveryReason = "review_findings"
	RecoveryReasonDuplicateFindings RecoveryReason = "duplicate_findings"
	RecoveryReasonStoryFailure      RecoveryReason = "story_failure"
	RecoveryReasonManualContinue    RecoveryReason = "manual_continue"
)

type RecoveryFinding struct {
	Category string `json:"category"`
	Path     string `json:"path"`
	Line     int    `json:"line,omitempty"`
	Summary  string `json:"summary"`
}

func RecoverFromFailure(
	codebaseContext, prdFile string,
	reason RecoveryReason,
	attempt, maxAttempts int,
	errMsg string,
	findings []RecoveryFinding,
	changedFiles []string,
) string {
	findingsSection := ""
	if len(findings) > 0 {
		data, _ := json.Marshal(findings)
		findingsSection = fmt.Sprintf(`
REVIEW FINDINGS (fix every item):
%s
`, string(data))
	}

	errorSection := ""
	if errMsg != "" {
		errorSection = fmt.Sprintf(`
FAILURE:
%s
`, errMsg)
	}

	escalation := ""
	if reason == RecoveryReasonDuplicateFindings || attempt > 1 {
		escalation = `
Previous recovery attempts did not resolve the issue. Try a different approach than before.
Remove incorrect generated artifacts instead of adding parallel files.
Prefer fixing integration paths over duplicating framework outputs.
`
	}

	return fmt.Sprintf(`You are %s, working inside the user's git repo on the feature branch.
%s%s%s%s
RECOVERY REASON: %s (attempt %d of %d)

Fix the problems above so Ralph can continue implementation review and testing.
- Address every review finding and failure cause directly in the repo.
- Delete stray generated files that are not part of the project's normal build pipeline.
- Do not mark PRD stories complete; Ralph handles story state.
- Run only the tests needed to validate your fixes.
- Stop when fixes are in place and tests are green.

PRD file: %s`,
		RecoveryAgentMarker,
		codebaseContextSection(codebaseContext),
		errorSection,
		findingsSection,
		escalation,
		reason,
		attempt,
		maxAttempts,
		prdFile,
	)
}

func IsRecoveryPrompt(prompt string) bool {
	return strings.Contains(prompt, RecoveryAgentMarker)
}
