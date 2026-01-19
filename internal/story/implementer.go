package story

import (
	"context"
	"fmt"
	"strings"

	"ralph/internal/config"
	"ralph/internal/git"
	"ralph/internal/prd"
	"ralph/internal/prompt"
	"ralph/internal/runner"
)

// Implementer handles story implementation
type Implementer struct {
	cfg    *config.Config
	runner *runner.Runner
	git    *git.Manager
}

// NewImplementer creates a new story implementer
func NewImplementer(cfg *config.Config) *Implementer {
	return &Implementer{
		cfg:    cfg,
		runner: runner.New(cfg),
		git:    git.New(),
	}
}

// Implement executes the implementation of a single story
func (i *Implementer) Implement(ctx context.Context, story *prd.Story, iteration int, p *prd.PRD, outputCh chan<- runner.OutputLine) (bool, error) {
	implPrompt := prompt.StoryImplementation(
		story.Title,
		story.Description,
		story.AcceptanceCriteria,
		story.TestSpec,
		story.ID,
		iteration,
		p.CompletedCount(),
		len(p.Stories),
	)

	result, err := i.runner.RunOpenCode(ctx, implPrompt, outputCh)
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
		if outputCh != nil {
			outputCh <- runner.OutputLine{Text: fmt.Sprintf("Warning: commit failed: %v", err), IsErr: true}
		}
	}

	return true, nil
}
