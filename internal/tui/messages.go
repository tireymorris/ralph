package tui

import (
	"ralph/internal/prompt"
	"ralph/internal/shared/prd"
)

type (
	phaseChangeMsg Phase

	resumeStartMsg struct {
		phase Phase
		prd   *prd.PRD
	}

	operationErrorMsg struct {
		err error
	}

	clarifyQuestionsMsg struct {
		questions []string
		answersCh chan<- []prompt.QuestionAnswer
	}
)
