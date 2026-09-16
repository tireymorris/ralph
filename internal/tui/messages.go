package tui

import (
	"github.com/tireymorris/ralph/internal/prompt"
	"github.com/tireymorris/ralph/internal/shared/prd"
	"github.com/tireymorris/ralph/internal/shared/session"
)

type (
	phaseChangeMsg Phase

	resumeStartMsg struct {
		phase    Phase
		prd      *prd.PRD
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
