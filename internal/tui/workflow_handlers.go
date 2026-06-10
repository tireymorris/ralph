package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/prompt"
	"ralph/internal/shared/runner"
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
	m.logger.AddLog("Clarifications received, updating PRD...")
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
		m.revisingPRD = false
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

	case events.EventPRDRevising:
		m.phase = PhasePRDGeneration
		m.revisingPRD = true
		m.logger.AddLog("Applying critique to PRD...")
		m.markMainScrollJump()

	case events.EventPRDReview:
		m.phase = PhasePRDReview
		m.revisingPRD = false
		m.prd = e.PRD
		m.logger.AddLog("PRD ready for review")
		m.markMainScrollJump()

	case events.EventStoryStarted:
		m.currentStory = e.Story
		m.phase = PhaseImplementation
		if m.prd == nil {
			if p, err := m.operationManager.PRDForImplementation(m.cfg); err != nil {
				m.logger.AddLog(fmt.Sprintf("Failed to load PRD: %v", err))
			} else {
				m.prd = p
			}
		}
		m.logger.AddLog(fmt.Sprintf("Starting: %s", e.Story.Title))
		m.markMainScrollJump()

	case events.EventStoryCompleted:
		m.logger.AddLog(fmt.Sprintf("Completed: %s", e.Story.Title))
		if m.prd != nil {
			if s := m.prd.GetStory(e.Story.ID); s != nil {
				s.Passes = true
			}
		}

	case events.EventImplementationReviewStarted:
		m.logger.AddLog(fmt.Sprintf("Implementation review started (iteration %d)", e.Iteration))
		m.markMainScrollJump()

	case events.EventImplementationReview:
		for _, f := range e.Findings {
			if f.Summary != "" {
				m.logger.AddLog(fmt.Sprintf("Review finding: %s", f.Summary))
			}
		}
		m.markMainScrollJump()

	case events.EventImplementationReviewCompleted:
		outcome := "clean"
		if !e.Clean {
			outcome = "findings"
		}
		m.logger.AddLog(fmt.Sprintf("Implementation review completed (iteration %d, %s)", e.Iteration, outcome))
		m.markMainScrollJump()

	case events.EventRecoveryStarted:
		m.phase = PhaseImplementation
		m.logger.AddLog(fmt.Sprintf("Recovery started (%s, attempt %d/%d)", e.Reason, e.Attempt, e.Max))
		m.markMainScrollJump()

	case events.EventRecoveryCompleted:
		outcome := "failed"
		if e.Success {
			outcome = "succeeded"
		}
		m.logger.AddLog(fmt.Sprintf("Recovery %s (%s, attempt %d)", outcome, e.Reason, e.Attempt))
		m.markMainScrollJump()

	case events.EventCleanupStarted:
		m.phase = PhaseCleanup
		m.logger.AddLog("Running post-implementation cleanup...")
		m.markMainScrollJump()

	case events.EventCleanupCompleted:
		m.logger.AddLog("Cleanup complete!")
		m.markMainScrollJump()

	case events.EventOutput:
		if !e.Verbose || m.verbose {
			m.logger.AddOutputLine(runner.OutputLine{Text: e.Text, IsErr: e.IsErr})
		}

	case events.EventError:
		m.logger.AddLog(fmt.Sprintf("Error: %v", e.Err))
		m.retryImplementation = m.phase == PhaseImplementation
		m.revisingPRD = false
		m.err = e.Err
		m.phase = PhaseFailed
		m.markMainScrollJump()

	case events.EventCompleted:
		m.phase = PhaseCompleted
		m.logger.AddLog("All stories completed!")
		m.markMainScrollJump()
	}

	return nil
}
