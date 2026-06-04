package tui

import (
	"ralph/internal/prompt"
)

type (
	phaseChangeMsg Phase

	operationErrorMsg struct {
		err error
	}

	clarifyQuestionsMsg struct {
		questions []string
		answersCh chan<- []prompt.QuestionAnswer
	}
)
