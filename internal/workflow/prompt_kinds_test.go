package workflow

import promptpkg "ralph/internal/prompt"

func isDiffReviewPrompt(p string) bool {
	return promptpkg.HasKind(p, promptpkg.KindDiffReview)
}

func isStoryImplementPrompt(p string) bool {
	return promptpkg.HasKind(p, promptpkg.KindStoryImplement)
}

func isCleanupPrompt(p string) bool {
	return promptpkg.HasKind(p, promptpkg.KindCleanup)
}

func isRecoveryPrompt(p string) bool {
	return promptpkg.HasKind(p, promptpkg.KindRecovery)
}
