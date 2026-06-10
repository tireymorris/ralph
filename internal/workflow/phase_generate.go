package workflow

import (
	"context"
	"fmt"

	"ralph/internal/prompt"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/logger"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
)

func (e *Executor) RunGenerate(ctx context.Context, userPrompt string) (*prd.PRD, error) {
	return e.RunGenerateWithAnswers(ctx, userPrompt, nil)
}

func (e *Executor) RunGenerateWithAnswers(ctx context.Context, userPrompt string, qas []prompt.QuestionAnswer) (*prd.PRD, error) {
	logger.Debug("generating PRD", "prompt_length", len(userPrompt))
	e.emit(EventPRDGenerating{})

	hasSource := workdirContainsSource(e.cfg.WorkDir)
	if !hasSource {
		logger.Info("working directory has no source code, treating as new project", "work_dir", e.cfg.WorkDir)
		e.emit(EventOutput{Output: Output{Text: "Warning: Working directory appears to have no source code. PRD will be generated for a new project."}})
	}

	e.emit(EventOutput{Output: Output{Text: "Analyzing codebase and generating PRD..."}})

	outputCh := make(chan runner.OutputLine, constants.EventChannelBuffer)
	go e.forwardOutput(outputCh)

	prdPrompt := prompt.PRDGenerationWithAnswers(userPrompt, e.cfg.PRDFile, "feature", !hasSource, qas)
	err := e.runner.Run(ctx, prdPrompt, outputCh)
	close(outputCh)

	if err != nil {
		logger.Error("PRD generation failed", "error", err)
		e.emit(EventError{Err: fmt.Errorf("PRD generation failed with runner %s: %w", e.cfg.Runner, err)})
		return nil, fmt.Errorf("PRD generation failed with runner %s: %w", e.cfg.Runner, err)
	}

	exists, err := e.store.Exists(e.cfg)
	if err != nil {
		logger.Error("failed to check for generated PRD", "error", err)
		e.emit(EventError{Err: fmt.Errorf("checking for generated PRD %s: %w", e.cfg.PRDFile, err)})
		return nil, fmt.Errorf("checking for generated PRD %s: %w", e.cfg.PRDFile, err)
	}
	if !exists {
		err := fmt.Errorf("AI completed but did not generate %s — it may not have understood the request", e.cfg.PRDFile)
		logger.Error("AI did not generate PRD file", "file", e.cfg.PRDFile)
		e.emit(EventError{Err: err})
		return nil, err
	}

	p, err := e.store.Load(e.cfg)
	if err != nil {
		logger.Error("failed to load generated PRD", "error", err)
		e.emit(EventError{Err: fmt.Errorf("failed to load generated PRD %s: %w", e.cfg.PRDFile, err)})
		return nil, fmt.Errorf("failed to load generated PRD %s: %w", e.cfg.PRDFile, err)
	}

	p, err = e.runPRDSelfReview(ctx, userPrompt)
	if err != nil {
		logger.Error("PRD self-review failed", "error", err)
		e.emit(EventError{Err: fmt.Errorf("PRD self-review failed: %w", err)})
		return nil, fmt.Errorf("PRD self-review failed: %w", err)
	}

	logger.Debug("PRD generated", "project", p.ProjectName, "stories", len(p.Stories))
	e.emit(EventPRDGenerated{PRD: p})
	e.emit(EventPRDReview{PRD: p})
	return p, nil
}
