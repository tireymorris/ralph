package prompt

import (
	"fmt"
	"strings"
)

func PRDGeneration(userPrompt, prdFile, branchPrefix string) string {
	return fmt.Sprintf(`You are Ralph, an autonomous software development agent. Implement: %s

PROCESS:
1. Analyze the codebase thoroughly - list files, read configs, understand the tech stack
2. Plan the implementation based on existing patterns
3. Write a PRD file to %s with user stories

REQUIRED OUTPUT - Write this exact JSON structure to %s:
{
  "project_name": "descriptive name",
  "branch_name": "%s/descriptive-branch-name",
  "stories": [
    {
      "id": "story-1",
      "title": "Short title",
      "description": "Detailed implementation requirements",
      "acceptance_criteria": ["criterion 1", "criterion 2"],
      "test_spec": "How to test this story",
      "priority": 1,
      "passes": false,
      "retry_count": 0
    }
  ]
}

REQUIREMENTS:
- Each story must be implementable in one iteration
- Include test_spec with specific test guidance for each story
- Set priority based on dependencies (1 = implement first)
- Create the git branch specified in branch_name
- Stories should build on each other logically

Write the PRD file now.`, userPrompt, prdFile, prdFile, branchPrefix)
}

func StoryImplementation(storyID, title, description string, acceptanceCriteria []string, testSpec string, prdFile string, iteration, completed, total int) string {
	if testSpec == "" {
		testSpec = "Create and run appropriate tests"
	}

	return fmt.Sprintf(`Implement story "%s" (ID: %s)

STORY DETAILS:
- Title: %s
- Description: %s
- Acceptance Criteria: %s
- Test Guidance: %s

CONTEXT: Iteration %d, %d/%d stories completed

PROCESS:
1. Find the relevant code for this feature
2. Implement the feature completely
3. Write tests following existing project patterns
4. Run tests and ensure they pass
5. Commit changes with message: "feat: %s"
6. Update %s - set "passes": true for story "%s"

CRITICAL: You MUST update %s to mark the story as complete when done.
If tests fail, increment "retry_count" in %s for this story instead.

Implement now.`,
		title, storyID,
		title,
		description,
		strings.Join(acceptanceCriteria, "; "),
		testSpec,
		iteration, completed, total,
		title,
		prdFile, storyID,
		prdFile,
		prdFile,
	)
}
