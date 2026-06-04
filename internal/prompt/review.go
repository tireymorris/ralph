package prompt

import "fmt"

func CriticalDiffReview(codebaseContext, prdFile string, changedFiles []string) string {
	return fmt.Sprintf(`You are Ralph's critical diff review agent, working inside the user's git repo on the feature branch.
%s%s
Review the changes in the listed files against the PRD in %s. Report issues as structured findings.

PRD file: %s`, codebaseContextSection(codebaseContext), changedFilesSection(changedFiles), prdFile, prdFile)
}
