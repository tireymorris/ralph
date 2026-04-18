package workflow

import (
	"context"
	"encoding/json"

	"ralph/internal/constants"
	"ralph/internal/logger"
	"ralph/internal/prompt"
	"ralph/internal/runner"
)

// RunClarify asks the AI to generate clarifying questions about the user's
// prompt, then emits EventClarifyingQuestions and blocks until the consumer
// sends answers back. Returns the answers (possibly empty if skipped).
func (e *Executor) RunClarify(ctx context.Context, userPrompt string) ([]prompt.QuestionAnswer, error) {
	isEmpty := isEmptyCodebase(e.cfg.WorkDir)

	e.emit(EventOutput{Output: Output{Text: "Analyzing request and generating clarifying questions..."}})

	outputCh := make(chan runner.OutputLine, constants.EventChannelBuffer)
	go e.forwardOutput(outputCh)

	clarifyPrompt := prompt.ClarifyingQuestions(userPrompt, ClarifyingQuestionsFile, isEmpty)
	err := e.runner.Run(ctx, clarifyPrompt, outputCh)
	close(outputCh)

	if err != nil {
		logger.Warn("clarifying questions generation failed, proceeding without", "error", err)
		return nil, nil
	}

	data, readErr := QuestionsFileReader{WorkDir: e.cfg.WorkDir}.ReadRemove()

	if readErr != nil {
		logger.Warn("AI did not write questions file, proceeding without clarification", "file", ClarifyingQuestionsFile)
		return nil, nil
	}

	var questions []string
	if jsonErr := json.Unmarshal(data, &questions); jsonErr != nil {
		logger.Warn("failed to parse questions file, proceeding without clarification", "error", jsonErr)
		return nil, nil
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
