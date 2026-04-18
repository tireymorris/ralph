package tui

import (
	"ralph/internal/prd"
	"ralph/internal/prompt"
)

type (
	prdGeneratedMsg struct{ prd *prd.PRD }
	prdErrorMsg     struct{ err error }
	phaseChangeMsg  Phase

	clarifyQuestionsMsg struct {
		questions []string
		answersCh chan<- []prompt.QuestionAnswer
	}
)
