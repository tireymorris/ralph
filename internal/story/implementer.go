package story

import (
	"context"
	"fmt"
	"strings"

	"ralph/internal/config"
	"ralph/internal/git"
	"ralph/internal/logger"
	"ralph/internal/prd"
	"ralph/internal/prompt"
	"ralph/internal/runner"
)

type GitCommitter interface {
	CommitStory(storyID, title, description string) error
}

type Implementer struct {
	cfg    *config.Config
	runner runner.CodeRunner
	git    GitCommitter
}

func NewImplementer(cfg *config.Config) *Implementer {
	return &Implementer{
		cfg:    cfg,
		runner: runner.New(cfg),
		git:    git.NewWithWorkDir(cfg.WorkDir),
	}
}

func NewImplementerWithDeps(cfg *config.Config, r runner.CodeRunner, g GitCommitter) *Implementer {
	return &Implementer{
		cfg:    cfg,
		runner: r,
		git:    g,
	}
}

func (i *Implementer) Implement(ctx context.Context, story *prd.Story, iteration int, p *prd.PRD, outputCh chan<- runner.OutputLine) (bool, error) {
	logger.Debug("implementing story",
		"story_id", story.ID,
		"title", story.Title,
		"iteration", iteration,
		"completed", p.CompletedCount(),
		"total", len(p.Stories))

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
		logger.Error("opencode run failed for story", "story_id", story.ID, "error", err)
		return false, fmt.Errorf("failed to run opencode: %w", err)
	}

	if result.Error != nil {
		logger.Debug("opencode returned error for story", "story_id", story.ID, "error", result.Error)
		return false, nil
	}

	if !strings.Contains(result.Output, "COMPLETED:") {
		logger.Debug("story not marked as completed", "story_id", story.ID)
		return false, nil
	}

	logger.Debug("story marked as completed, committing", "story_id", story.ID)
	if err := i.git.CommitStory(story.ID, story.Title, story.Description); err != nil {
		logger.Warn("commit failed for story", "story_id", story.ID, "error", err)
		if outputCh != nil {
			outputCh <- runner.OutputLine{Text: fmt.Sprintf("Warning: commit failed: %v", err), IsErr: true}
		}
	}

	return true, nil
}
