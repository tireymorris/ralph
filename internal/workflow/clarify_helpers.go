package workflow

import (
	"context"
	"encoding/json"
	"fmt"

	"ralph/internal/prompt"
	"ralph/internal/shared/logger"
)

func (e *Executor) runClarifyRunner(ctx context.Context, userPrompt string, isEmpty bool) error {
	clarifyPrompt := prompt.ClarifyingQuestions(userPrompt, ClarifyingQuestionsFile, isEmpty)
	return e.runWithForwardedOutput(ctx, clarifyPrompt)
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
