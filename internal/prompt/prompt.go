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

func PRDValidation(prdJSON string) string {
	return fmt.Sprintf(`Analyze this PRD for actionability and technical correctness:

PRD:
%s

CODEBASE CONTEXT:
- Go 1.24.0 with standard testing
- Ralph: autonomous dev agent with Bubbletea TUI, structured logging
- Key directories: internal/workflow (orchestration), internal/runner (AI CLI execution), internal/prd (PRD models), internal/prompt (templates)
- Current complexity: 650-word PRD generation prompt, 350-word story prompts, 8 event types, JSON repair mechanism
- Testing: go test ./..., go test -race ./...

VALIDATION REQUIREMENTS:
1. **Specificity Analysis**: For each story, identify:
   - Vague terms ("simplify", "optimize", "reduce", "improve") without quantifiable metrics
   - Missing technical details (file paths, function names, line counts)
   - Non-testable acceptance criteria

2. **Technical Accuracy**: 
   - Verify proposed changes match actual codebase structure
   - Check dependencies and assumptions are correct
   - Ensure implementation scope is realistic

3. **Actionability Criteria**:
   - Each story must have specific, measurable requirements
   - Acceptance criteria must be verifiable
   - No ambiguous technical instructions

FIXES REQUIRED:
- Replace vague terms with specific metrics (e.g., "reduce prompt from 650 to 200 words")
- Add concrete technical details (file paths, function signatures)
- Make acceptance criteria testable
- Ensure stories are implementation-ready

Output the improved PRD as valid JSON.`, prdJSON)
}

func StoryImplementation(storyID, title, description string, acceptanceCriteria []string, featureTestSpec, context, prdFile string, iteration, completed, total int) string {
	return fmt.Sprintf(`Implement story: %s

Description: %s
Done when: %s

Steps:
1. Implement using codebase context
2. Add tests if functionality is complete  
3. Run existing tests
4. Commit changes: "feat: %s"
5. Update %s - set "passes": true for story "%s"

Progress: Iteration %d, %d/%d stories completed`,
		title,
		description,
		strings.Join(acceptanceCriteria, "; "),
		title,
		prdFile, storyID,
		iteration, completed, total,
	)
}
