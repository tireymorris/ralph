package prompt

import (
	"fmt"
	"strings"
)

func PRDGeneration(userPrompt string) string {
	return fmt.Sprintf(`You are Ralph, an autonomous software development agent. Your task is to implement: %s

Follow this process:

1. COMPREHENSIVE PROJECT ANALYSIS
   - First, list all files in the current directory to understand the project structure
   - Read key configuration and documentation files: README.md, package.json, go.mod, Cargo.toml, pyproject.toml, requirements.txt, etc.
   - Identify the technology stack by examining file extensions, build files, and dependencies
   - Find and read the main entry point files (main.go, app.py, index.js, etc.)
   - Locate and examine existing test files to understand testing framework, naming conventions, and patterns
   - Search for existing source code files to understand code organization and patterns
   - Note any build scripts, CI/CD configurations, and deployment setups
   - Analyze the existing codebase thoroughly to understand how features are implemented

2. DETAILED IMPLEMENTATION PLANNING
   - Based on the codebase analysis, create a detailed plan for implementing the requested feature
   - Identify which existing files need to be modified or extended
   - Determine what new files need to be created
   - Consider dependencies, imports, and integration points
   - Plan the implementation order to ensure each story builds on previous ones

3. CREATE PRD
   - Generate comprehensive, actionable user stories based on the thorough codebase analysis
   - Each story must be implementable in one iteration and should leverage existing patterns
   - Include detailed acceptance criteria that can be verified
   - Set priorities based on dependencies and logical implementation order (1=highest)
   - CRITICAL: Each story MUST include a specific test_spec with guidance for writing runtime tests

4. TEST SPECIFICATION REQUIREMENTS
   - The test_spec field provides GUIDANCE for writing actual test code for the specific feature
   - An actual test file will be created and run for EACH story before moving to the next
   - Tests must validate RUNTIME behavior, not just compilation
   - IMPORTANT: Tests should follow existing project conventions (file location, naming, framework)
   - IMPORTANT: Name tests after the FEATURE being tested (e.g., "user_authentication_test.go" not "integration_test.go")
   - For UI features: describe interactions to automate (clicks, inputs, assertions on DOM)
   - For API integrations: describe requests to make and expected responses
   - For setup stories: describe how to verify the setup works (e.g., app starts, imports work)
   - Include specific assertions that can be coded (e.g., "response status should be 200", "element with class X should contain Y")
   - Each test builds on previous tests - later stories should verify previous functionality still works

5. OUTPUT REQUIREMENTS
   - Respond ONLY with raw JSON (no markdown, no explanation)

Required JSON format:
{
  "project_name": "descriptive project name",
  "branch_name": "feature/branch-name",
  "stories": [
    {
      "id": "story-1",
      "title": "Story title",
      "description": "Detailed description based on codebase analysis",
      "acceptance_criteria": ["criterion 1", "criterion 2"],
      "test_spec": "Test guidance: 1) Perform specific action, 2) Assert expected behavior, 3) Verify integration points.",
      "priority": 1,
      "passes": false
    }
  ]
}

CRITICAL:
- Perform thorough codebase exploration before generating the PRD
- Ensure stories are based on actual project structure and existing patterns
- Return only the JSON object, nothing else.
- Every story MUST have a non-empty test_spec field with actionable, specific test guidance.
- Test specs should be detailed enough to write and run automated tests.
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
