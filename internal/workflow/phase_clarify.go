package workflow

import (
	"context"

	"ralph/internal/prompt"
	"ralph/internal/shared/logger"
)

// RunClarify emits questions and waits for consumer answers; skipped questions return nil.
func (e *Executor) RunClarify(ctx context.Context, userPrompt string) ([]prompt.QuestionAnswer, error) {
	if e.cfg.AutoApprove {
		e.emit(EventOutput{Output: Output{Text: "AutoApprove: skipping clarification"}})
		return nil, nil
	}

	hasSource := workdirContainsSource(e.cfg.WorkDir)

	e.emit(EventOutput{Output: Output{Text: "Analyzing request and generating clarifying questions..."}})

	if err := e.runClarifyRunner(ctx, userPrompt, !hasSource); err != nil {
		logger.Warn("clarifying questions generation failed, proceeding without", "error", err)
		return nil, nil
	}

	data, readErr := QuestionsFileReader{WorkDir: e.cfg.WorkDir}.ReadRemove()
	if readErr != nil {
		return continueWithoutClarification("AI did not write questions file, proceeding without clarification", "file", ClarifyingQuestionsFile)
	}

	questions, jsonErr := parseClarifyingQuestions(data)
	if jsonErr != nil {
		return continueWithoutClarification("failed to parse questions file, proceeding without clarification", "error", jsonErr)
	}

	if len(questions) == 0 {
		logger.Debug("AI generated no clarifying questions")
		return nil, nil
	}

	if e.eventsCh == nil {
		logger.Debug("no event channel, skipping clarifying questions")
		return nil, nil
	}
	answersCh := make(chan []prompt.QuestionAnswer, 1)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case e.eventsCh <- EventClarifyingQuestions{Questions: questions, AnswersCh: answersCh}:
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case answers := <-answersCh:
		return answers, nil
	}
}
