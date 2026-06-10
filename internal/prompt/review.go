package prompt

import "fmt"

func CriticalDiffReview(codebaseContext, prdFile string, changedFiles []string) string {
	return fmt.Sprintf(`You are Ralph's critical diff review agent, working inside the user's git repo on the feature branch.
%s%s
Review the changes in the listed files against the PRD in %s. The changed-files list includes committed branch deltas and uncommitted working-tree edits; read the current file contents on disk. Do not report incomplete delivery solely because work is uncommitted when the listed files already contain the implementation.

Report issues as structured findings.

PRD file: %s

OUTPUT FORMAT:
When finished, write findings between these markers as a JSON array. Use [] when there are no issues.

===ralph-findings===
[{"category":"<category>","path":"<file path>","summary":"<one-line description>"}]
===/ralph-findings===

Each object must include category, path, and summary. Include line only when the issue is tied to a specific line number.`, codebaseContextSection(codebaseContext), changedFilesSection(changedFiles), prdFile, prdFile)
}
