package prompt

func CriticalDiffReview(codebaseContext, prdFile string, changedFiles []string) string {
	return mustRender("diff-review", DiffReviewData{
		Context:      codebaseContext,
		PRDFile:      prdFile,
		ChangedFiles: changedFiles,
	})
}
