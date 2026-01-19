package prd

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"ralph/internal/config"
	"ralph/internal/prompt"
	"ralph/internal/runner"
)

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

func (g *Generator) Generate(ctx context.Context, userPrompt string, outputCh chan<- runner.OutputLine) (*PRD, error) {
	prdPrompt := prompt.PRDGeneration(userPrompt)

	result, err := g.runner.RunOpenCode(ctx, prdPrompt, outputCh)
	if err != nil {
		return nil, fmt.Errorf("failed to run opencode: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("opencode error: %w", result.Error)
	}

	p, err := parseResponse(result.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PRD: %w", err)
	}

	if err := validate(p); err != nil {
		return nil, fmt.Errorf("invalid PRD: %w", err)
	}

	sort.Slice(p.Stories, func(i, j int) bool {
		return p.Stories[i].Priority < p.Stories[j].Priority
	})

	return p, nil
}

func parseResponse(response string) (*PRD, error) {
	response = strings.TrimSpace(response)

	start := strings.Index(response, "{")
	if start == -1 {
		return nil, fmt.Errorf("no JSON object found in response")
	}

	end := findMatchingBrace(response, start)
	if end == -1 {
		return nil, fmt.Errorf("no complete JSON object found in response")
	}

	var p PRD
	if err := json.Unmarshal([]byte(response[start:end]), &p); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &p, nil
}

func findMatchingBrace(s string, start int) int {
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
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

	for i, story := range p.Stories {
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
