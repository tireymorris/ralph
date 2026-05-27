package workflow

import (
	"context"
	"fmt"

	"ralph/internal/prompt"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/logger"
	"ralph/internal/shared/runner"
)

// RunCritiqueRevision applies user critique to the PRD, re-runs clarification, then returns to review.
func (e *Executor) RunCritiqueRevision(ctx context.Context, userPrompt, critique string) error {
	e.emit(EventPRDRevising{})

	if err := e.applyCritique(ctx, userPrompt, critique); err != nil {
		return err
	}

	qas, err := e.RunClarify(ctx, userPrompt)
	if err != nil {
		return err
	}

	if len(qas) > 0 {
		if err := e.applyClarifications(ctx, userPrompt, qas); err != nil {
			return err
		}
	}

	p, err := e.store.Load(e.cfg)
	if err != nil {
		logger.Error("failed to load PRD after critique revision", "error", err)
		wrappedErr := fmt.Errorf("failed to load PRD %s after critique revision: %w", e.cfg.PRDFile, err)
		e.emit(EventError{Err: wrappedErr})
		return wrappedErr
	}

	e.emit(EventPRDReview{PRD: p})
	return nil
}

func (e *Executor) applyCritique(ctx context.Context, userPrompt, critique string) error {
	logger.Debug("applying critique to PRD", "critique_length", len(critique))
	e.emit(EventOutput{Output: Output{Text: "Researching and applying critique to PRD..."}})

	outputCh := make(chan runner.OutputLine, constants.EventChannelBuffer)
	go e.forwardOutput(outputCh)

	revisionPrompt := prompt.PRDCritiqueRevision(userPrompt, e.cfg.PRDFile, critique)
	err := e.runner.Run(ctx, revisionPrompt, outputCh)
	close(outputCh)

	if err != nil {
		logger.Error("PRD critique revision failed", "error", err)
		e.emit(EventError{Err: fmt.Errorf("PRD critique revision failed with runner %s: %w", e.cfg.Runner, err)})
		return fmt.Errorf("PRD critique revision failed with runner %s: %w", e.cfg.Runner, err)
	}

	return e.ensurePRDExists("critique revision")
}

func (e *Executor) applyClarifications(ctx context.Context, userPrompt string, qas []prompt.QuestionAnswer) error {
	logger.Debug("applying post-critique clarifications to PRD", "answers", len(qas))
	e.emit(EventOutput{Output: Output{Text: "Applying clarifications to revised PRD..."}})

	outputCh := make(chan runner.OutputLine, constants.EventChannelBuffer)
	go e.forwardOutput(outputCh)

	revisionPrompt := prompt.PRDClarificationRevision(userPrompt, e.cfg.PRDFile, qas)
	err := e.runner.Run(ctx, revisionPrompt, outputCh)
	close(outputCh)

	if err != nil {
		logger.Error("PRD clarification revision failed", "error", err)
		e.emit(EventError{Err: fmt.Errorf("PRD clarification revision failed with runner %s: %w", e.cfg.Runner, err)})
		return fmt.Errorf("PRD clarification revision failed with runner %s: %w", e.cfg.Runner, err)
	}

	return e.ensurePRDExists("clarification revision")
}

func (e *Executor) ensurePRDExists(phase string) error {
	exists, err := e.store.Exists(e.cfg)
	if err != nil {
		logger.Error("failed to check PRD after revision", "error", err, "phase", phase)
		wrappedErr := fmt.Errorf("checking for PRD %s after %s: %w", e.cfg.PRDFile, phase, err)
		e.emit(EventError{Err: wrappedErr})
		return wrappedErr
	}
	if !exists {
		err := fmt.Errorf("AI completed %s but did not update %s", phase, e.cfg.PRDFile)
		logger.Error("AI did not update PRD file", "file", e.cfg.PRDFile, "phase", phase)
		e.emit(EventError{Err: err})
		return err
	}
	return nil
}
