package workflow

import (
	"context"
	"encoding/json"
	"fmt"

	"ralph/internal/shared/constants"
	"ralph/internal/logger"
	"ralph/internal/prompt"
	"ralph/internal/runner"
)

func (e *Executor) runClarifyRunner(ctx context.Context, userPrompt string, isEmpty bool) error {
	outputCh := make(chan runner.OutputLine, constants.EventChannelBuffer)
	go e.forwardOutput(outputCh)

	clarifyPrompt := prompt.ClarifyingQuestions(userPrompt, ClarifyingQuestionsFile, isEmpty)
	err := e.runner.Run(ctx, clarifyPrompt, outputCh)
	close(outputCh)
	return err
}

func parseClarifyingQuestions(data []byte) ([]string, error) {
	var questions []string
	if err := json.Unmarshal(data, &questions); err != nil {
		return nil, fmt.Errorf("parse questions file: %w", err)
	}
	return questions, nil
}

func continueWithoutClarification(reason string, args ...any) ([]prompt.QuestionAnswer, error) {
	logger.Warn(reason, args...)
	return nil, nil
}
