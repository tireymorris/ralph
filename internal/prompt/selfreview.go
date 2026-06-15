package prompt

// PRDSelfReviewVerdictFile is the temporary JSON file the AI writes its self-review verdict to.
const PRDSelfReviewVerdictFile = ".ralph/prd_review.json"

// PRDSelfReview instructs the agent to critique and revise the PRD in place, then write a verdict file.
func PRDSelfReview(userPrompt, prdFile string, round, maxRounds int) string {
	return mustRender("prd-self-review", PRDSelfReviewData{
		UserPrompt:  userPrompt,
		PRDFile:     prdFile,
		Round:       round,
		MaxRounds:   maxRounds,
		VerdictFile: PRDSelfReviewVerdictFile,
	})
}
