package story

import (
	"context"
	"fmt"
	"strings"

	"ralph/internal/config"
	"ralph/internal/git"
	"ralph/internal/prd"
	"ralph/internal/runner"
)

// Implementer handles story implementation
type Implementer struct {
	cfg    *config.Config
	runner *runner.Runner
	git    *git.Manager
}

func NewImplementer(cfg *config.Config) *Implementer {
	return &Implementer{
		cfg:    cfg,
		runner: runner.New(cfg),
		git:    git.New(),
	}
}

// Implement executes the implementation of a single story
func (i *Implementer) Implement(ctx context.Context, story *prd.Story, iteration int, p *prd.PRD, outputCh chan<- runner.OutputLine) (bool, error) {
	completed := p.CompletedCount()
	total := len(p.Stories)

	prompt := buildImplementationPrompt(story, iteration, completed, total)

	result, err := i.runner.RunOpenCode(ctx, prompt, outputCh)
	if err != nil {
		return false, fmt.Errorf("failed to run opencode: %w", err)
	}

	if result.Error != nil {
		return false, nil
	}

	// Check if implementation was successful
	if !strings.Contains(result.Output, "COMPLETED:") {
		return false, nil
	}

	// Commit changes
	if err := i.git.CommitStory(story.ID, story.Title, story.Description); err != nil {
		// Log but don't fail - the implementation succeeded
		if outputCh != nil {
			outputCh <- runner.OutputLine{Text: fmt.Sprintf("Warning: commit failed: %v", err), IsErr: true}
		}
	}

	return true, nil
}

func buildImplementationPrompt(story *prd.Story, iteration, completed, total int) string {
	testSpec := story.TestSpec
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
2. IMPLEMENT the feature completely
3. WRITE AN INTEGRATION TEST for this story:
   - Create/update test file: tests/%s.test.{js,ts,rb,py} (match project language)
   - Test MUST verify the feature works at RUNTIME, not just compilation
   - Use appropriate testing framework (Playwright, Puppeteer, Vitest, Jest, RSpec, pytest, etc.)
4. RUN THE TEST and ensure it PASSES - do NOT proceed until tests pass
5. RUN ALL PREVIOUS TESTS to ensure no regressions
6. COMMIT changes including both implementation and test files

CRITICAL REQUIREMENTS:
- You MUST write an actual test file, not just describe tests
- You MUST run the test and see it pass in the output
- Do NOT mark complete if you only ran lint/build - tests must pass
- The test must verify RUNTIME behavior (e.g., app starts, UI renders, API responds)

When the integration test passes and changes are committed, respond:
"COMPLETED: [summary] | TEST: [test file path] | RESULT: [pass/fail with brief output]"

CRITICAL: Respond ONLY with the completion message, nothing else.`,
		story.Title,
		story.Description,
		strings.Join(story.AcceptanceCriteria, ", "),
		testSpec,
		iteration,
		completed,
		total,
		story.ID,
	)
}
