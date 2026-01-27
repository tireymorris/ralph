package prompt

import (
	"fmt"
	"strings"
)

func PRDGeneration(userPrompt, prdFile, branchPrefix string) string {
	return fmt.Sprintf(`Create a PRD for: %s

Write JSON to %s:
{
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

Requirements:
- context: essential technical info for implementation 
- test_spec: STRING with 3-5 holistic test scenarios (NOT array)
- stories: implementation steps with specific, measurable requirements
- acceptance_criteria: verifiable completion conditions
- Priority: based on dependencies (1 = first)

Task: Analyze codebase, create branch, write PRD file, STOP.`, userPrompt, prdFile, branchPrefix)
}

func PRDValidation(prdJSON, prdFile, codebaseContext string) string {
	contextSection := ""
	if codebaseContext != "" {
		contextSection = fmt.Sprintf(`
CODEBASE CONTEXT:
%s
`, codebaseContext)
	}

	return fmt.Sprintf(`Analyze this PRD for actionability and technical correctness:

PRD:
%s
%s
VALIDATION REQUIREMENTS:
1. Each story must have specific, measurable requirements
2. Acceptance criteria must be verifiable (not vague)
3. Vague terms ("simplify", "optimize", "reduce", "improve") need quantifiable metrics
4. Technical details (file paths, function names) should be present where relevant

FIXES REQUIRED:
- Replace vague terms with specific metrics
- Add concrete technical details where missing
- Make acceptance criteria testable
- Ensure stories are implementation-ready

Write the improved PRD as valid JSON to %s.`, prdJSON, contextSection, prdFile)
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
