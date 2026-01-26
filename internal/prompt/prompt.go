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
  "test_spec": "Holistic test specification (see below)",
  "stories": [
    {
      "id": "story-1",
      "title": "Short title",
      "description": "What to implement",
      "acceptance_criteria": ["done when X works"],
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

TEST_SPEC FIELD REQUIREMENTS:
The "test_spec" field defines how to test the ENTIRE feature holistically:
- Focus on end-to-end behavior, not individual story testing
- Define 3-5 key test scenarios that verify the feature works correctly
- Include both happy path and important edge cases
- Tests should be written once and cover the full implementation
- Do NOT create separate test suites per story - one cohesive test approach

STORY REQUIREMENTS:
- Stories are implementation steps, NOT separate features requiring separate tests
- Keep stories small and focused on implementation milestones
- acceptance_criteria should be 1-2 simple "done when" conditions
- Set priority based on dependencies (1 = implement first)
- Create the git branch specified in branch_name
- Stories should build on each other logically

Write the PRD file now.`, userPrompt, prdFile, prdFile, branchPrefix)
}

func JSONRepair(prdFile, parseError string) string {
	return fmt.Sprintf(`The file %s contains invalid JSON that failed to parse.

Error: %s

Read %s, find and fix the JSON syntax error, then save the corrected file.

Common issues to check:
- Missing or extra commas between fields/array elements
- Misplaced brackets ] or braces }
- Unclosed strings
- Trailing commas before closing brackets

Fix the JSON syntax error while preserving all the data. Do not change any content, only fix the syntax.`,
		prdFile, parseError, prdFile)
}

func StoryImplementation(storyID, title, description string, acceptanceCriteria []string, featureTestSpec, context, prdFile string, iteration, completed, total int) string {
	contextSection := ""
	if context != "" {
		contextSection = fmt.Sprintf(`
CODEBASE CONTEXT:
%s

`, context)
	}

	testSection := ""
	if featureTestSpec != "" {
		testSection = fmt.Sprintf(`
FEATURE TEST SPEC (holistic - covers all stories):
%s

`, featureTestSpec)
	}

	return fmt.Sprintf(`Implement story "%s" (ID: %s)
%s%s
STORY DETAILS:
- Title: %s
- Description: %s
- Done when: %s

PROGRESS: Iteration %d, %d/%d stories completed

PROCESS:
1. Implement the story using the codebase context above
2. If this story completes testable functionality, add tests per the feature test spec
3. Run existing tests to ensure nothing is broken
4. Commit changes with message: "feat: %s"
5. Update %s - set "passes": true for story "%s"

TESTING GUIDANCE:
- Tests should cover the feature holistically, not each story individually
- Only add new tests when this story completes functionality worth testing
- Avoid duplicating test coverage across stories

CONTEXT UPDATES (IMPORTANT):
After completing this story, you MUST update the "context" field in %s if you:
- Created new modules, classes, or utilities that future stories should reuse
- Established patterns (e.g., validation approach, error handling style)
- Added helpers, mixins, or shared functionality
- Discovered project conventions not previously documented
Append new information to the existing context. This helps future stories avoid re-reading code.

CRITICAL: You MUST update %s to mark the story as complete when done.
If tests fail, increment "retry_count" in %s for this story instead.

Implement now.`,
		title, storyID,
		contextSection,
		testSection,
		title,
		description,
		strings.Join(acceptanceCriteria, "; "),
		iteration, completed, total,
		title,
		prdFile, storyID,
		prdFile,
		prdFile,
		prdFile,
	)
}
