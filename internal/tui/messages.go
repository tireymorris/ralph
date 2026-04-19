package tui

import (
	"ralph/internal/prompt"
)

type (
	phaseChangeMsg Phase

	clarifyQuestionsMsg struct {
		questions []string
		answersCh chan<- []prompt.QuestionAnswer
	}
)
