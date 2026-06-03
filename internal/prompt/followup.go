package prompt

import "fmt"

func FollowUpRevision(userMessage, prdFile, runTranscript string) string {
	return fmt.Sprintf(`You are Ralph's planning agent, working inside the user's git repo on the feature branch.

The user completed a prior run and requested a follow-up change:
%s

Read %s and update the PRD JSON to incorporate this follow-up. Preserve existing story IDs where stories remain relevant; add or adjust stories as needed. Increment planning detail in descriptions and acceptance criteria where the change requires it.

Run transcript (recent events from the prior run):
%s

Task:
1. Read %s and analyze the codebase enough to address the follow-up — budget ~5–15 targeted file reads.
2. Update %s to reflect the follow-up request.
3. Write the updated PRD file, then STOP — do not implement anything.`, userMessage, prdFile, runTranscript, prdFile, prdFile)
}
