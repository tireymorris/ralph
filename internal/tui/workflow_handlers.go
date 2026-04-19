package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/prompt"
	"ralph/internal/runner"
	"ralph/internal/workflow/events"
)

func (m *Model) buildAnswers() []prompt.QuestionAnswer {
	if len(m.clarifyQuestions) == 0 {
		return nil
	}
	qas := make([]prompt.QuestionAnswer, len(m.clarifyQuestions))
	for i, q := range m.clarifyQuestions {
		answer := ""
		if i < len(m.clarifyInputs) {
			answer = m.clarifyInputs[i].Value()
		}
		qas[i] = prompt.QuestionAnswer{Question: q, Answer: answer}
	}
	return qas
}

func (m *Model) submitClarifyingAnswers(qas []prompt.QuestionAnswer) []tea.Cmd {
	if m.clarifyAnswersCh != nil {
		m.clarifyAnswersCh <- qas
		m.clarifyAnswersCh = nil
	}
	m.phase = PhasePRDGeneration
	m.logger.AddLog("Clarifications received, generating PRD...")
	return []tea.Cmd{m.operationManager.ListenForEvents()}
}

func (m *Model) handleWorkflowEvent(event events.Event) tea.Cmd {
	switch e := event.(type) {
	case events.EventClarifyingQuestions:
		return func() tea.Msg {
			return clarifyQuestionsMsg{
				questions: e.Questions,
				answersCh: e.AnswersCh,
			}
		}

	case events.EventPRDGenerating:
		m.phase = PhasePRDGeneration
		m.logger.AddLog("Generating PRD...")
		m.markMainScrollJump()

	case events.EventPRDGenerated:
		m.prd = e.PRD
		m.logger.AddLog(fmt.Sprintf("PRD generated: %s (%d stories)", e.PRD.ProjectName, len(e.PRD.Stories)))
		if m.dryRun {
			m.phase = PhaseCompleted
			m.logger.AddLog("Dry run complete - PRD saved to " + m.cfg.PRDFile)
		} else {
			m.phase = PhasePRDReview
		}
		m.markMainScrollJump()

	case events.EventPRDLoaded:
		m.prd = e.PRD
		m.logger.AddLog(fmt.Sprintf("Loaded PRD: %s (%d/%d completed)",
			e.PRD.ProjectName, e.PRD.CompletedCount(), len(e.PRD.Stories)))
		if m.dryRun {
			m.phase = PhaseCompleted
		} else {
			m.phase = PhasePRDReview
		}
		m.markMainScrollJump()

	case events.EventPRDReview:
		m.phase = PhasePRDReview
		m.prd = e.PRD
		m.logger.AddLog("PRD ready for review")
		m.markMainScrollJump()

	case events.EventStoryStarted:
		m.currentStory = e.Story
		m.logger.AddLog(fmt.Sprintf("Starting: %s", e.Story.Title))

	case events.EventStoryCompleted:
		if e.Success {
			m.logger.AddLog(fmt.Sprintf("Completed: %s", e.Story.Title))
			if m.prd != nil {
				if s := m.prd.GetStory(e.Story.ID); s != nil {
					s.Passes = true
				}
			}
		} else {
			m.logger.AddLog(fmt.Sprintf("Retrying: %s", e.Story.Title))
		}

	case events.EventOutput:
		if !e.Verbose || m.verbose {
			m.logger.AddOutputLine(runner.OutputLine{Text: e.Text, IsErr: e.IsErr})
		}

	case events.EventError:
		m.logger.AddLog(fmt.Sprintf("Error: %v", e.Err))
		m.err = e.Err
		m.markMainScrollJump()

	case events.EventCompleted:
		m.phase = PhaseCompleted
		m.logger.AddLog("All stories completed!")
		m.markMainScrollJump()
	}

	return nil
}
