package prompt

import (
	"fmt"
	"strings"
)

func PRDGeneration(userPrompt, prdFile, branchPrefix string, isEmptyCodebase bool) string {
	contextGuidance := `- context: describe ONLY the tech stack and patterns you ACTUALLY observe in the codebase`
	if isEmptyCodebase {
		contextGuidance = `Note: The working directory has no existing source code. This is a new project.
- context: describe ONLY the tech stack specified in the user's request, or state "New project - no existing codebase"
- Do NOT assume or invent a tech stack the user did not mention`
	}

	return fmt.Sprintf(`Create a PRD for: %s

Write JSON to %s:
{
  "version": 1,
  "project_name": "descriptive name",
  "branch_name": "%s/descriptive-branch-name",
  "context": "Tech stack, project structure, testing approach, key patterns",
  "test_spec": "String describing 3-5 holistic test scenarios for the entire feature",
  "stories": [
    {
      "id": "story-1",
      "title": "Short title",
      "description": "Specific implementation task with technical details",
      "acceptance_criteria": ["testable completion condition"],
      "priority": 1,
      "passes": false,
      "retry_count": 0
    }
  ]
}

CRITICAL QUALITY REQUIREMENTS:
%s
- test_spec: STRING with 3-5 holistic test scenarios (NOT array)
- stories: implementation steps with specific, measurable requirements
- acceptance_criteria: MUST be verifiable and specific. NEVER use vague terms without quantification:
  * Avoid vague verbs without metrics: "simplify", "optimize", "reduce", "improve", "enhance", "streamline", "refactor"
  * If using these verbs, add quantifiable metrics (e.g., "reduce from 650 to 200 words", "optimize query to <100ms")
  * Avoid vague adjectives without specifics: "proper", "appropriate", "comprehensive", "good", "correct", "consistent", "clean", "robust"
  * Replace vague criteria like "proper error handling" with testable conditions like "returns 400 status with error message for invalid input"
- description: Include concrete technical details (file paths, function names, API endpoints) where relevant
- Priority: based on dependencies (1 = first)

Each story must be implementation-ready with specific, measurable requirements that can be verified through testing or code inspection.

Task: Analyze codebase, create branch, write high-quality PRD file, STOP.`, userPrompt, prdFile, branchPrefix, contextGuidance)
}

func StoryImplementation(storyID, title, description string, acceptanceCriteria []string, featureTestSpec, codebaseContext, prdFile string, iteration, completed, total int) string {
	contextSection := ""
	if codebaseContext != "" {
		contextSection = fmt.Sprintf(`
CODEBASE CONTEXT:
%s
`, codebaseContext)
	}

	testSection := ""
	if featureTestSpec != "" {
		testSection = fmt.Sprintf(`
FEATURE TEST SPEC:
%s
`, featureTestSpec)
	}

	return fmt.Sprintf(`Implement story: %s
%s%s
Description: %s
Done when: %s

Steps:
1. Implement using codebase context above
2. Add tests per feature test spec if this story completes testable functionality
3. Run existing tests
4. Commit changes: "feat: %s"
5. Update %s - set "passes": true for story "%s"
6. If tests fail, increment "retry_count" for this story in %s instead

After completing this story, update the "context" field in %s if you created new modules, established patterns, or discovered conventions that future stories should know about.

Progress: Iteration %d, %d/%d stories completed`,
		title,
		contextSection,
		testSection,
		description,
		strings.Join(acceptanceCriteria, "; "),
		title,
		prdFile, storyID,
		prdFile,
		prdFile,
		iteration, completed, total,
	)
}
