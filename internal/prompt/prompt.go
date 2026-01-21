package prompt

import (
	"fmt"
	"strings"
)

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
   - CRITICAL: Each story MUST include a test_spec with guidance for writing tests for that specific feature
   
3. TEST SPECIFICATION REQUIREMENTS
   - The test_spec field provides GUIDANCE for writing test code for the specific feature
   - An actual test file will be created and run for EACH story before moving to the next
   - Tests must validate RUNTIME behavior, not just compilation
   - IMPORTANT: Tests should follow existing project conventions (file location, naming, framework)
   - IMPORTANT: Name tests after the FEATURE being tested (e.g., "entity_discovery_tool_spec.rb" not "integration_spec.rb")
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
      "test_spec": "Test guidance: 1) Call the feature method, 2) Assert expected return value, 3) Verify side effects if any.",
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

1. FIND THE RELEVANT CODE - search for the specific feature/module named in the story title
   - Search for the CLASS, MODULE, or FUNCTION name from the story (e.g., "EntityDiscoveryTool", "UserService")
   - Find existing test files for THAT SPECIFIC module (e.g., entity_discovery_tool_spec.rb, user_service_test.py)
   - Do NOT search for generic terms like "integration" or "test" - search for the FEATURE NAME
2. READ the existing code and its test file to understand patterns
   - Note the test framework, naming conventions, and test structure
3. IMPLEMENT the feature completely
4. WRITE A TEST for this story:
   - Put the test in the SAME test file as other tests for this module (or create one following naming conventions)
   - Test MUST verify the feature works at RUNTIME, not just compilation
   - Use descriptive test names that match project conventions
5. RUN THE TEST and ensure it PASSES - do NOT proceed until tests pass
6. RUN ALL PREVIOUS TESTS to ensure no regressions
7. COMMIT changes including both implementation and test files

CRITICAL REQUIREMENTS:
- Search for the SPECIFIC FEATURE NAME in the story, not generic terms
- You MUST write an actual test file, not just describe tests
- You MUST run the test and see it pass in the output
- Do NOT mark complete if you only ran lint/build - tests must pass
- The test must verify RUNTIME behavior (e.g., app starts, UI renders, API responds)
- FOLLOW EXISTING TEST PATTERNS - do not create story-X.test.{ext} files

When the test passes and changes are committed, respond:
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
