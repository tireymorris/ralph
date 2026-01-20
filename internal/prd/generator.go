package prd

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"ralph/internal/config"
	"ralph/internal/errors"
	"ralph/internal/logger"
	"ralph/internal/prompt"
	"ralph/internal/runner"
)

type Generator struct {
	cfg    *config.Config
	runner runner.CodeRunner
}

func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{
		cfg:    cfg,
		runner: runner.New(cfg),
	}
}

func NewGeneratorWithRunner(cfg *config.Config, r runner.CodeRunner) *Generator {
	return &Generator{
		cfg:    cfg,
		runner: r,
	}
}

func (g *Generator) Generate(ctx context.Context, userPrompt string, outputCh chan<- runner.OutputLine) (*PRD, error) {
	logger.Debug("generating PRD prompt", "user_prompt_length", len(userPrompt))
	prdPrompt := prompt.PRDGeneration(userPrompt)

	result, err := g.runner.RunOpenCode(ctx, prdPrompt, outputCh)
	if err != nil {
		logger.Error("opencode run failed", "error", err)
		return nil, errors.OpencodeError{Op: "execution", Err: err}
	}

	if result.Error != nil {
		logger.Error("opencode returned error", "error", result.Error)
		return nil, errors.OpencodeError{Op: "execution", Err: result.Error}
	}

	logger.Debug("parsing PRD response", "response_length", len(result.Output))
	p, err := parseResponse(result.Output)
	if err != nil {
		logger.Error("failed to parse PRD response", "error", err)
		return nil, err
	}

	if err := validate(p); err != nil {
		logger.Error("PRD validation failed", "error", err)
		return nil, errors.PRDError{Op: "validation", Err: err}
	}

	sort.Slice(p.Stories, func(i, j int) bool {
		return p.Stories[i].Priority < p.Stories[j].Priority
	})

	logger.Debug("PRD generated successfully",
		"project", p.ProjectName,
		"stories", len(p.Stories),
		"branch", p.BranchName)

	return p, nil
}

func parseResponse(response string) (*PRD, error) {
	response = strings.TrimSpace(response)

	start := strings.Index(response, "{")
	if start == -1 {
		return nil, errors.PRDError{Op: "parsing", Err: fmt.Errorf("no JSON object found in response")}
	}

	// Use json.Decoder to properly parse JSON, handling all edge cases
	// including braces inside quoted strings
	decoder := json.NewDecoder(strings.NewReader(response[start:]))

	var p PRD
	if err := decoder.Decode(&p); err != nil {
		// If streaming decode fails, try to extract JSON manually
		// with proper string handling as a fallback
		end := findMatchingBrace(response, start)
		if end == -1 {
			return nil, errors.PRDError{Op: "parsing", Err: fmt.Errorf("no complete JSON object found in response: %w", err)}
		}

		if err := json.Unmarshal([]byte(response[start:end]), &p); err != nil {
			return nil, errors.PRDError{Op: "parsing", Err: fmt.Errorf("failed to parse JSON: %w", err)}
		}
	}

	return &p, nil
}

// findMatchingBrace finds the closing brace for a JSON object starting at 'start'.
// It properly handles braces inside quoted strings and escape sequences.
func findMatchingBrace(s string, start int) int {
	depth := 0
	inString := false
	escaped := false

	for i := start; i < len(s); i++ {
		ch := s[i]

		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' && inString {
			escaped = true
			continue
		}

		if ch == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		switch ch {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}
	return -1
}

func validate(p *PRD) error {
	if p.ProjectName == "" {
		return fmt.Errorf("missing project_name")
	}

	if len(p.Stories) == 0 {
		return fmt.Errorf("no stories defined")
	}

	// Check for duplicate story IDs
	seenIDs := make(map[string]bool)

	for i, story := range p.Stories {
		if story.ID == "" {
			return fmt.Errorf("story %d missing id", i+1)
		}
		if seenIDs[story.ID] {
			return fmt.Errorf("duplicate story id: %s", story.ID)
		}
		seenIDs[story.ID] = true

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
