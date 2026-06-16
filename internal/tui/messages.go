package tui

import (
	"ralph/internal/prompt"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/session"
)

type (
	phaseChangeMsg Phase

	resumeStartMsg struct {
		phase Phase
		prd   *prd.PRD
		snapshot session.RunSnapshot
	}

	operationErrorMsg struct {
		err error
	}

	clarifyQuestionsMsg struct {
		questions []string
		answersCh chan<- []prompt.QuestionAnswer
	}
)
