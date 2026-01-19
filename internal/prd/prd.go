package prd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"ralph/internal/config"
	"ralph/internal/runner"
)

// Generator creates PRDs from prompts
type Generator struct {
	cfg    *config.Config
	runner *runner.Runner
}

func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{
		cfg:    cfg,
		runner: runner.New(cfg),
	}
}

// Generate creates a PRD from the given prompt
func (g *Generator) Generate(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) (*PRD, error) {
	prdPrompt := buildPRDPrompt(prompt)

	result, err := g.runner.RunOpenCode(ctx, prdPrompt, outputCh)
	if err != nil {
		return nil, fmt.Errorf("failed to run opencode: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("opencode error: %w", result.Error)
	}

	prd, err := parseResponse(result.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PRD: %w", err)
	}

	if err := validatePRD(prd); err != nil {
		return nil, fmt.Errorf("invalid PRD: %w", err)
	}

	// Sort stories by priority
	sort.Slice(prd.Stories, func(i, j int) bool {
		return prd.Stories[i].Priority < prd.Stories[j].Priority
	})

	return prd, nil
}

// Load reads a PRD from the configured file
func Load(cfg *config.Config) (*PRD, error) {
	data, err := os.ReadFile(cfg.PRDFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read PRD file: %w", err)
	}

	var prd PRD
	if err := json.Unmarshal(data, &prd); err != nil {
		return nil, fmt.Errorf("failed to parse PRD: %w", err)
	}

	// Initialize retry counts if not present
	for _, story := range prd.Stories {
		if story.RetryCount == 0 && !story.Passes {
			story.RetryCount = 0
		}
	}

	return &prd, nil
}

// Save writes the PRD to the configured file
func Save(cfg *config.Config, prd *PRD) error {
	data, err := json.MarshalIndent(prd, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal PRD: %w", err)
	}

	if err := os.WriteFile(cfg.PRDFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write PRD file: %w", err)
	}

	return nil
}

// Delete removes the PRD file
func Delete(cfg *config.Config) error {
	if _, err := os.Stat(cfg.PRDFile); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(cfg.PRDFile)
}

// Exists checks if a PRD file exists
func Exists(cfg *config.Config) bool {
	_, err := os.Stat(cfg.PRDFile)
	return err == nil
}

func buildPRDPrompt(prompt string) string {
	return fmt.Sprintf(`You are Ralph, an autonomous software development agent. Your task is to implement: %s

Follow this process:

1. PROJECT ANALYSIS
   - Scan current directory to understand existing codebase
   - Identify technology stack, patterns, conventions
   - Note dependencies, tests, build setup
   
2. CREATE PRD
   - Generate comprehensive user stories
   - Each story must be implementable in one iteration
   - Include acceptance criteria and priorities (1=highest)
   - CRITICAL: Each story MUST include a test_spec with guidance for writing integration tests
   
3. TEST SPECIFICATION REQUIREMENTS
   - The test_spec field provides GUIDANCE for writing actual integration test code
   - An actual test file will be created and run for EACH story before moving to the next
   - Tests must validate RUNTIME behavior, not just compilation
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
- Tests are cumulative - each story's test should also verify previous stories still work.`, prompt)
}

func parseResponse(response string) (*PRD, error) {
	// Try to find JSON in the response
	response = strings.TrimSpace(response)

	// Look for JSON object
	start := strings.Index(response, "{")
	if start == -1 {
		return nil, fmt.Errorf("no JSON object found in response")
	}

	// Find the matching closing brace
	depth := 0
	end := -1
	for i := start; i < len(response); i++ {
		switch response[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = i + 1
				break
			}
		}
		if end > 0 {
			break
		}
	}

	if end == -1 {
		return nil, fmt.Errorf("no complete JSON object found in response")
	}

	jsonStr := response[start:end]

	var prd PRD
	if err := json.Unmarshal([]byte(jsonStr), &prd); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &prd, nil
}

func validatePRD(prd *PRD) error {
	if prd.ProjectName == "" {
		return fmt.Errorf("missing project_name")
	}

	if len(prd.Stories) == 0 {
		return fmt.Errorf("no stories defined")
	}

	for i, story := range prd.Stories {
		if story.ID == "" {
			return fmt.Errorf("story %d missing id", i+1)
		}
		if story.Title == "" {
			return fmt.Errorf("story %d missing title", i+1)
		}
		if story.Description == "" {
			return fmt.Errorf("story %d missing description", i+1)
		}
		if len(story.AcceptanceCriteria) == 0 {
			return fmt.Errorf("story %d missing acceptance_criteria", i+1)
		}
		if story.Priority == 0 {
			return fmt.Errorf("story %d missing priority", i+1)
		}
	}

	return nil
}

// NextPendingStory returns the next story to implement
func (p *PRD) NextPendingStory(maxRetries int) *Story {
	var best *Story
	for _, story := range p.Stories {
		if story.Passes {
			continue
		}
		if story.RetryCount >= maxRetries {
			continue
		}
		if best == nil || story.Priority < best.Priority {
			best = story
		}
	}
	return best
}

// CompletedCount returns the number of completed stories
func (p *PRD) CompletedCount() int {
	count := 0
	for _, story := range p.Stories {
		if story.Passes {
			count++
		}
	}
	return count
}

// FailedStories returns stories that have exceeded retry limits
func (p *PRD) FailedStories(maxRetries int) []*Story {
	var failed []*Story
	for _, story := range p.Stories {
		if !story.Passes && story.RetryCount >= maxRetries {
			failed = append(failed, story)
		}
	}
	return failed
}

// AllCompleted returns true if all stories are done
func (p *PRD) AllCompleted() bool {
	for _, story := range p.Stories {
		if !story.Passes {
			return false
		}
	}
	return true
}
