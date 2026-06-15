package prompt

import (
	"encoding/json"
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
	findingsJSON := ""
	if len(findings) > 0 {
		data, _ := json.Marshal(findings)
		findingsJSON = string(data)
	}
	escalate := reason == RecoveryReasonDuplicateFindings || attempt > 1
	return mustRender("recovery", RecoveryData{
		AgentMarker:  RecoveryAgentMarker,
		Context:      codebaseContext,
		ErrorMessage: errMsg,
		FindingsJSON: findingsJSON,
		Escalate:     escalate,
		Reason:       reason,
		Attempt:      attempt,
		MaxAttempts:  maxAttempts,
		PRDFile:      prdFile,
	})
}

func IsRecoveryPrompt(prompt string) bool {
	return strings.Contains(prompt, RecoveryAgentMarker)
}
