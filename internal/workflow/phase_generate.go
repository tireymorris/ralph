package workflow

import (
	"context"
	"fmt"

	"ralph/internal/constants"
	"ralph/internal/logger"
	"ralph/internal/prd"
	"ralph/internal/prompt"
	"ralph/internal/runner"
)

func (e *Executor) RunGenerate(ctx context.Context, userPrompt string) (*prd.PRD, error) {
	return e.RunGenerateWithAnswers(ctx, userPrompt, nil)
}

func (e *Executor) RunGenerateWithAnswers(ctx context.Context, userPrompt string, qas []prompt.QuestionAnswer) (*prd.PRD, error) {
	logger.Debug("generating PRD", "prompt_length", len(userPrompt))
	e.emit(EventPRDGenerating{})

	isEmpty := isEmptyCodebase(e.cfg.WorkDir)
	if isEmpty {
		logger.Info("working directory has no source code, treating as new project", "work_dir", e.cfg.WorkDir)
		e.emit(EventOutput{Output: Output{Text: "Warning: Working directory appears to have no source code. PRD will be generated for a new project."}})
	}

	e.emit(EventOutput{Output: Output{Text: "Analyzing codebase and generating PRD..."}})

	outputCh := make(chan runner.OutputLine, constants.EventChannelBuffer)
	go e.forwardOutput(outputCh)

	prdPrompt := prompt.PRDGenerationWithAnswers(userPrompt, e.cfg.PRDFile, "feature", isEmpty, qas)
	err := e.runner.Run(ctx, prdPrompt, outputCh)
	close(outputCh)

	if err != nil {
		logger.Error("PRD generation failed", "error", err)
		e.emit(EventError{Err: fmt.Errorf("PRD generation failed with model %s: %w", e.cfg.Model, err)})
		return nil, fmt.Errorf("PRD generation failed with model %s: %w", e.cfg.Model, err)
	}

	if !e.store.Exists(e.cfg) {
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

	logger.Debug("PRD generated", "project", p.ProjectName, "stories", len(p.Stories))
	e.emit(EventPRDGenerated{PRD: p})
	e.emit(EventPRDReview{PRD: p})
	return p, nil
}
