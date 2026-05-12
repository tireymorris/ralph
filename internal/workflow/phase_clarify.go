package workflow

import (
	"context"

	"ralph/internal/logger"
	"ralph/internal/prompt"
)

// RunClarify asks the AI to generate clarifying questions about the user's
// prompt, then emits EventClarifyingQuestions and blocks until the consumer
// sends answers back. Returns the answers (possibly empty if skipped).
func (e *Executor) RunClarify(ctx context.Context, userPrompt string) ([]prompt.QuestionAnswer, error) {
	isEmpty := isEmptyCodebase(e.cfg.WorkDir)

	e.emit(EventOutput{Output: Output{Text: "Analyzing request and generating clarifying questions..."}})

	if err := e.runClarifyRunner(ctx, userPrompt, isEmpty); err != nil {
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
