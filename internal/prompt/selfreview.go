package prompt

import "fmt"

// PRDSelfReviewVerdictFile is the temporary JSON file the AI writes its self-review verdict to.
const PRDSelfReviewVerdictFile = ".ralph/prd_review.json"

// PRDSelfReview instructs the agent to critique and revise the PRD in place, then write a verdict file.
func PRDSelfReview(userPrompt, prdFile string, round, maxRounds int) string {
	return fmt.Sprintf(`You are Ralph's planning agent, working inside the user's git repo on the feature branch.

The user's original request was: %s

This is PRD self-review round %d of %d.

Critically review the PRD in %s against the actual codebase and revise it in place wherever it falls short of this rubric:
- Every acceptance criterion must be objectively verifiable: an exact command, file path, event name, or observable behavior — never subjective adjectives like "proper", "clean", or "robust".
- Every file, function, and symbol a story references must exist in the repo unless the story itself creates it. Check the repo — do not trust the PRD.
- Each story must be sized for one focused, additive diff with no drive-by refactors and no scope creep.
- Every acceptance criterion must be writable as failing tests first (TDD).
- "depends_on" must be minimal and correct — only dependencies a story genuinely cannot start without.
- "context" must let an implementer start without re-discovering the architecture.
- Prefer the approach with the fewest touched lines.

Task:
1. Read %s and inspect the codebase enough to verify every file, function, and symbol the stories reference.
2. Revise %s in place wherever it falls short of the rubric. Preserve existing story IDs and "passes" values for stories that are unchanged.
3. Write your verdict as JSON to %s:
{"approved": <true only if the PRD now fully satisfies the rubric>, "summary": "<one or two sentences describing what you changed, or why you approved>"}
4. STOP — do not implement anything.`, userPrompt, round, maxRounds, prdFile, prdFile, prdFile, PRDSelfReviewVerdictFile)
}
