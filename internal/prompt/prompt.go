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
3. Write a PRD file to %s with user stories and captured context

REQUIRED OUTPUT - Write this exact JSON structure to %s:
{
  "project_name": "descriptive name",
  "branch_name": "%s/descriptive-branch-name",
  "context": "Captured codebase context (see below)",
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

CONTEXT FIELD REQUIREMENTS:
The "context" field must capture everything a developer needs to work on this codebase:
- Language/framework and versions (e.g., "Ruby 3.2 with RSpec", "Go 1.21 with standard testing")
- Project structure (key directories and their purpose)
- Important files to reference (main entry points, config files, existing patterns)
- Testing approach (test framework, where tests live, how to run them)
- Naming conventions and code style patterns observed
- Any existing utilities, helpers, or abstractions that should be reused
Keep it concise but complete - this context will be provided to agents implementing each story.

STORY REQUIREMENTS:
- Each story must be implementable in one iteration
- Include test_spec with specific test guidance for each story
- Set priority based on dependencies (1 = implement first)
- Create the git branch specified in branch_name
- Stories should build on each other logically

Write the PRD file now.`, userPrompt, prdFile, prdFile, branchPrefix)
}

func StoryImplementation(storyID, title, description string, acceptanceCriteria []string, testSpec, context, prdFile string, iteration, completed, total int) string {
	if testSpec == "" {
		testSpec = "Create and run appropriate tests"
	}

	contextSection := ""
	if context != "" {
		contextSection = fmt.Sprintf(`
CODEBASE CONTEXT:
%s

`, context)
	}

	return fmt.Sprintf(`Implement story "%s" (ID: %s)
%s
STORY DETAILS:
- Title: %s
- Description: %s
- Acceptance Criteria: %s
- Test Guidance: %s

PROGRESS: Iteration %d, %d/%d stories completed

PROCESS:
1. Implement the feature completely using the codebase context above
2. Write tests following existing project patterns
3. Run tests and ensure they pass
4. Commit changes with message: "feat: %s"
5. Update %s - set "passes": true for story "%s"
6. If you created new utilities, patterns, or discovered important info for future stories, update the "context" field in %s

CRITICAL: You MUST update %s to mark the story as complete when done.
If tests fail, increment "retry_count" in %s for this story instead.

Implement now.`,
		title, storyID,
		contextSection,
		title,
		description,
		strings.Join(acceptanceCriteria, "; "),
		testSpec,
		iteration, completed, total,
		title,
		prdFile, storyID,
		prdFile,
		prdFile,
		prdFile,
	)
}
