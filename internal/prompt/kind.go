package prompt

import "strings"

const (
	kindMarkerPrefix = "===ralph-prompt-kind:"
	kindMarkerSuffix = "==="
)

const (
	KindClarify                  = "clarify"
	KindPRDGenerate              = "prd-generate"
	KindPRDSelfReview            = "prd-self-review"
	KindPRDCritiqueRevision      = "prd-critique-revision"
	KindPRDClarificationRevision = "prd-clarification-revision"
	KindStoryImplement           = "story-implement"
	KindDiffReview               = "diff-review"
	KindRecovery                 = "recovery"
	KindCleanup                  = "cleanup"
	KindFollowUp                 = "followup"
)

func wrapWithKind(kind, body string) string {
	if kind == "" {
		return body
	}
	return kindMarkerPrefix + kind + kindMarkerSuffix + "\n" + body
}

func Kind(prompt string) string {
	start := strings.Index(prompt, kindMarkerPrefix)
	if start < 0 {
		return ""
	}
	rest := prompt[start+len(kindMarkerPrefix):]
	end := strings.Index(rest, kindMarkerSuffix)
	if end < 0 {
		return ""
	}
	return rest[:end]
}

func HasKind(prompt, kind string) bool {
	return Kind(prompt) == kind
}
