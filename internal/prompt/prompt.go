package prompt

import (
	"fmt"
	"strings"
)

// QuestionAnswer holds a clarifying question and the user's answer.
type QuestionAnswer struct {
	Question string
	Answer   string
}

// ClarifyingQuestions generates a prompt asking the AI to produce a JSON file
// with clarifying questions about the user's request. The AI writes a JSON
// array of question strings to questionsFile, then stops.
func ClarifyingQuestions(userPrompt, questionsFile string, isEmptyCodebase bool) string {
	codebaseNote := "an existing codebase"
	if isEmptyCodebase {
		codebaseNote = "a new project (no existing source code)"
	}

	return fmt.Sprintf(`You are Ralph's planning agent, working inside the user's git repo on %s.

The user's request is: %s

Before generating a full PRD, identify any ambiguities or missing details that would significantly affect how you design the solution. Write a JSON file at %s containing ONLY an array of clarifying question strings (no other keys):

["Question 1?", "Question 2?", ...]

Rules:
- Ask 0-5 concise, specific questions. Return [] if nothing is genuinely unclear — do not invent questions to fill a quota.
- Only ask about things that would change the technical approach
- Do NOT ask about things you can reasonably infer or decide yourself
- Do NOT ask for things already specified in the request
- Prefer questions about: scope boundaries, integration requirements, non-functional requirements (performance/scale), or preferred approaches when multiple are equally valid
- Write the JSON file, then STOP — do not implement anything`, codebaseNote, userPrompt, questionsFile)
}

func PRDGeneration(userPrompt, prdFile, branchPrefix string, isEmptyCodebase bool) string {
	return PRDGenerationWithAnswers(userPrompt, prdFile, branchPrefix, isEmptyCodebase, nil)
}

func PRDGenerationWithAnswers(userPrompt, prdFile, branchPrefix string, isEmptyCodebase bool, qas []QuestionAnswer) string {
	contextGuidance := `- context: describe ONLY the tech stack and patterns you ACTUALLY observe in the codebase`
	if isEmptyCodebase {
		contextGuidance = `Note: The working directory has no existing source code. This is a new project.
- context: describe ONLY the tech stack specified in the user's request, or state "New project - no existing codebase"
- Do NOT assume or invent a tech stack the user did not mention`
	}

	clarificationsSection := ""
	if len(qas) > 0 {
		var sb strings.Builder
		sb.WriteString("\nUSER CLARIFICATIONS:\n")
		for i, qa := range qas {
			sb.WriteString(fmt.Sprintf("Q%d: %s\nA%d: %s\n", i+1, qa.Question, i+1, qa.Answer))
		}
		clarificationsSection = sb.String()
	}

	return fmt.Sprintf(`You are Ralph's planning agent, working inside the user's git repo.

Create a PRD for: %s
%s
Write JSON to %s. Each field value below is a description of what to write — do not copy the descriptions literally.
{
  "version": 1,
  "project_name": <a short descriptive name for the feature>,
  "branch_name": "%s/<kebab-case-branch-name>",
  "context": <tech stack, project structure, testing approach (including the exact test runner command), and key patterns — concrete, not placeholder text>,
  "test_spec": <a single STRING (not an array) describing 3-5 holistic test scenarios for the whole feature>,
  "stories": [
    {
      "id": "story-1",
      "title": <short title>,
      "description": <specific implementation task with concrete technical details>,
      "acceptance_criteria": [<testable completion condition>],
      "priority": 1,
      "depends_on": [],
      "passes": false,
      "retry_count": 0
    }
  ]
}

CRITICAL QUALITY REQUIREMENTS:
%s
- test_spec: STRING with 3-5 holistic test scenarios (NOT an array)
- stories: size each story small enough to complete in roughly 3-10 red/green/commit TDD cycles. If a story feels larger than that, split it. Prefer many small stories over a few big ones.
- depends_on: ONLY include if this story genuinely cannot start until another is complete (e.g., "api-story" before "ui-story"). Most stories should have an empty array [].
- acceptance_criteria: MUST be verifiable and specific. NEVER use vague terms without quantification:
  * Avoid vague verbs without metrics: "simplify", "optimize", "reduce", "improve", "enhance", "streamline", "refactor"
  * If using these verbs, add quantifiable metrics (e.g., "reduce from 650 to 200 words", "optimize query to <100ms")
  * Avoid vague adjectives without specifics: "proper", "appropriate", "comprehensive", "good", "correct", "consistent", "clean", "robust"
  * Replace vague criteria like "proper error handling" with testable conditions like "returns 400 status with error message for invalid input"
- description: Include concrete technical details (file paths, function names, API endpoints) where relevant
- Priority: based on dependencies (1 = first)

Each story must be implementation-ready with specific, measurable requirements that can be verified through testing or code inspection.

Task:
1. Analyze the codebase so your "context" field captures real observed patterns and the actual test runner command.
2. Create and check out a new git branch named exactly the "branch_name" you chose (e.g., "git checkout -b %s/<your-branch-name>").
3. Write the PRD file, then STOP.`, userPrompt, clarificationsSection, prdFile, branchPrefix, contextGuidance, branchPrefix)
}

func StoryImplementation(storyID, title, description string, acceptanceCriteria []string, featureTestSpec, codebaseContext, prdFile string, iteration, completed, total int, dependsOn []string, parallelCount int) string {
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

	dependsSection := ""
	if len(dependsOn) > 0 {
		dependsSection = fmt.Sprintf(`
DEPENDENCIES: This story depends on: %s
Before starting, re-read %s and confirm each of those stories has "passes": true. If any dependency is not yet passing, stop and do not implement this story.`, strings.Join(dependsOn, ", "), prdFile)
	}

	parallelNote := ""
	if parallelCount > 1 {
		parallelNote = `
PARALLEL EXECUTION: Other stories are being worked on by peer agents at the same time. To minimize conflicts:
- Keep your diff narrow — touch the fewest files possible for each commit.
- Pull the latest state of any file just before you edit it; another story may have committed to it moments ago.
- Do NOT leave TODO comments telling other stories what to do. Coordinate only through the PRD file and real code.
- Write tests so they don't depend on execution order or shared mutable fixtures.`
	}

	return fmt.Sprintf(`You are Ralph's implementation agent, working inside the user's git repo on the feature branch.

Implement story: %s
%s%s%s%s
Description: %s
Done when: %s

Work in tight TDD cycles. Do NOT implement the whole story in one pass. Break it into the smallest independently testable slices you can, then for EACH slice:

  a. RED — write one failing test for the next small piece of behavior. Use the project's actual test runner (see the codebase context if provided); do not invent a new framework. Run it. Confirm it fails for the right reason.
  b. GREEN — write the minimum code needed to make that test pass. Run the new test. Run the rest of the existing tests to confirm no regressions.
  c. (optional) REFACTOR — clean up while tests stay green.
  d. COMMIT — commit just that slice on its own.

Commit message rules:
- One short sentence, lowercase, imperative mood, no trailing period.
- NO conventional-commit prefixes. Do NOT start with "feat:", "fix:", "refactor:", "chore:", "test:", "docs:", etc.
- Describe what this one slice does, not the whole story. Example: "parse empty input as zero-length token list", not "feat: add parser".

Repeat the red → green → commit loop until every acceptance criterion is satisfied. Many small commits per story is expected and preferred over one large commit.

IMPORTANT — failure semantics: If this story does not reach the done state (tests failing, blocked, etc.) before you stop, Ralph will ` + "`git reset --hard`" + ` back to the SHA before you started and retry the story from scratch. Your small commits will be wiped on failure. This is why each commit should be a real, tested step forward — not a WIP save.

When every acceptance criterion passes and the full test suite is green:
- Edit %s and set "passes": true for story "%s".
- Do NOT touch "retry_count"; Ralph manages that field.

After completing this story, update the "context" field in %s ONLY if you established a new pattern, added a new module, or discovered a convention that future stories need to know. Skip the context update for routine stories.

Progress: Iteration %d, %d/%d stories completed`,
		title,
		contextSection,
		testSection,
		dependsSection,
		parallelNote,
		description,
		strings.Join(acceptanceCriteria, "; "),
		prdFile, storyID,
		prdFile,
		iteration, completed, total,
	)
}
