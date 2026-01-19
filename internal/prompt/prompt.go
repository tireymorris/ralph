package prompt

import (
	"fmt"
	"strings"
)

// PRDGeneration returns the prompt for generating a PRD from user requirements
func PRDGeneration(userPrompt string) string {
	return fmt.Sprintf(`You are Ralph, an autonomous software development agent. Your task is to implement: %s

Follow this process:

1. PROJECT ANALYSIS
   - Scan current directory to understand existing codebase
   - Identify technology stack, patterns, conventions
   - Note dependencies, tests, build setup
   - IMPORTANT: Analyze existing test files to understand naming conventions and test patterns
   
2. CREATE PRD
   - Generate comprehensive user stories
   - Each story must be implementable in one iteration
   - Include acceptance criteria and priorities (1=highest)
   - CRITICAL: Each story MUST include a test_spec with guidance for writing integration tests
   
3. TEST SPECIFICATION REQUIREMENTS
   - The test_spec field provides GUIDANCE for writing actual integration test code
   - An actual test file will be created and run for EACH story before moving to the next
   - Tests must validate RUNTIME behavior, not just compilation
   - IMPORTANT: Tests should follow existing project conventions (file location, naming, framework)
   - For UI features: describe interactions to automate (clicks, inputs, assertions on DOM)
   - For API integrations: describe requests to make and expected responses
   - For setup stories: describe how to verify the setup works (e.g., app starts, imports work)
   - Include specific assertions that can be coded (e.g., "element with class X should contain Y")
   - Each test builds on previous tests - later stories should verify previous functionality still works
   
4. OUTPUT REQUIREMENTS
   - Respond ONLY with raw JSON (no markdown, no explanation)
   
Required JSON format:
{
  "project_name": "descriptive project name",
  "branch_name": "feature/branch-name",
  "stories": [
    {
      "id": "story-1",
      "title": "Story title",
      "description": "Detailed description",
      "acceptance_criteria": ["criterion 1", "criterion 2"],
      "test_spec": "Integration test guidance: 1) Start app, 2) Navigate to X, 3) Assert element Y is visible, 4) Click Z, 5) Assert result.",
      "priority": 1,
      "passes": false
    }
  ]
}

CRITICAL: 
- Return only the JSON object, nothing else.
- Every story MUST have a non-empty test_spec field with actionable test guidance.
- Test specs should be specific enough to write automated tests (selectors, expected values, actions).
- Tests are cumulative - each story's test should also verify previous stories still work.`, userPrompt)
}

// StoryImplementation returns the prompt for implementing a single story
func StoryImplementation(title, description string, acceptanceCriteria []string, testSpec string, iteration, completed, total int) string {
	if testSpec == "" {
		testSpec = "No test spec provided - create and run appropriate tests"
	}

	return fmt.Sprintf(`You are Ralph implementing story: %s

Story: %s
Acceptance Criteria: %s

Test Spec Guidelines:
%s

Context: Iteration %d (%d/%d stories done)

IMPLEMENTATION PROCESS:

1. READ existing code to understand patterns and test setup
   - Look for existing test files to understand naming conventions and test framework
   - Check for test directories, config files (jest.config, vitest.config, pytest.ini, etc.)
2. IMPLEMENT the feature completely
3. WRITE AN INTEGRATION TEST for this story:
   - FOLLOW EXISTING TEST CONVENTIONS in the project (file naming, location, framework)
   - If no tests exist, use standard conventions for the language/framework
   - Test MUST verify the feature works at RUNTIME, not just compilation
   - DO NOT use generic names like "story-1.test.js" - use descriptive names that match project conventions
   - Examples: "units.test.js", "test_temperature_conversion.py", "units_spec.rb"
4. RUN THE TEST and ensure it PASSES - do NOT proceed until tests pass
5. RUN ALL PREVIOUS TESTS to ensure no regressions
6. COMMIT changes including both implementation and test files

CRITICAL REQUIREMENTS:
- You MUST write an actual test file, not just describe tests
- You MUST run the test and see it pass in the output
- Do NOT mark complete if you only ran lint/build - tests must pass
- The test must verify RUNTIME behavior (e.g., app starts, UI renders, API responds)
- FOLLOW EXISTING TEST PATTERNS - do not create story-X.test.{ext} files

When the integration test passes and changes are committed, respond:
"COMPLETED: [summary] | TEST: [test file path] | RESULT: [pass/fail with brief output]"

CRITICAL: Respond ONLY with the completion message, nothing else.`,
		title,
		description,
		strings.Join(acceptanceCriteria, ", "),
		testSpec,
		iteration,
		completed,
		total,
	)
}
