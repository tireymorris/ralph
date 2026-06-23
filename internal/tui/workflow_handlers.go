package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/prompt"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
	"ralph/internal/shared/runstate"
	"ralph/internal/shared/session"
	"ralph/internal/workflow/events"
)

func (m *Model) syncPresentation(fallbackPhase string) {
	snapshot, loaded, err := m.operationManager.refreshPresentation(fallbackPhase)
	if err != nil {
		return
	}
	snapshot.Activity = m.activity
	m.prd = loaded
	m.snapshot = snapshot
	if snapshot.CurrentStory != nil {
		m.currentStory = snapshot.CurrentStory
	}
}

func (m *Model) activeStoryForActivity() (*prd.Story, string, string) {
	if m.currentStory != nil {
		return m.currentStory, m.currentStory.ID, m.currentStory.Title
	}
	if m.prd != nil {
		if story := m.prd.NextReadyStory(); story != nil {
			return story, story.ID, story.Title
		}
	}
	return nil, "", ""
}

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
		progress := e.PRD.RunProgress()
		m.logger.AddLog(fmt.Sprintf("PRD generated: %s (%d stories)", e.PRD.ProjectName, progress.Total))
		if m.dryRun {
			m.phase = PhaseCompleted
			m.logger.AddLog("Dry run complete - PRD saved to " + m.cfg.PRDFile)
		} else if m.cfg.AutoApprove {
			m.phase = PhasePRDGeneration
		} else {
			m.phase = PhasePRDReview
		}
		m.markMainScrollJump()

	case events.EventPRDLoaded:
		m.prd = e.PRD
		progress := e.PRD.RunProgress()
		m.logger.AddLog(fmt.Sprintf("Loaded PRD: %s (%d/%d completed)",
			e.PRD.ProjectName, progress.Completed, progress.Total))
		if m.dryRun {
			m.phase = PhaseCompleted
		} else if m.cfg.AutoApprove {
			m.phase = PhasePRDGeneration
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
		m.revisingPRD = false
		m.prd = e.PRD
		if m.cfg.AutoApprove {
			m.logger.AddLog("PRD auto-approved, continuing to implementation")
		} else {
			m.phase = PhasePRDReview
			m.logger.AddLog("PRD ready for review")
		}
		m.markMainScrollJump()

	case events.EventStoryStarted:
		m.currentStory = e.Story
		m.phase = PhaseImplementation
		_, storyID, storyTitle := m.activeStoryForActivity()
		if e.Story != nil {
			storyID = e.Story.ID
			storyTitle = e.Story.Title
		}
		m.activity = session.RunActivity{
			Kind:       session.ActivityImplementing,
			StoryID:    storyID,
			StoryTitle: storyTitle,
		}
		m.syncPresentation(runstate.PhaseImplement)
		m.logger.AddLog(fmt.Sprintf("Starting: %s", e.Story.Title))
		m.markMainScrollJump()

	case events.EventSliceStarted, events.EventSliceCompleted:
		m.syncPresentation(runstate.PhaseImplement)

	case events.EventStoryCompleted:
		m.logger.AddLog(fmt.Sprintf("Completed: %s", e.Story.Title))
		m.syncPresentation(runstate.PhaseImplement)
		m.markMainScrollJump()

	case events.EventImplementationReviewStarted:
		_, storyID, storyTitle := m.activeStoryForActivity()
		m.activity = session.RunActivity{
			Kind:       session.ActivityReview,
			StoryID:    storyID,
			StoryTitle: storyTitle,
			Iteration:  e.Iteration,
		}
		m.phase = PhaseCleanup
		m.logger.AddLog(fmt.Sprintf("Cleanup started (iteration %d)", e.Iteration))
		m.syncPresentation(runstate.PhaseCleanup)
		m.markMainScrollJump()

	case events.EventImplementationReview:
		m.activity.Kind = session.ActivityReview
		m.activity.FindingCount = len(e.Findings)
		for _, f := range e.Findings {
			if f.Summary != "" {
				m.logger.AddLog(fmt.Sprintf("Review finding: %s", f.Summary))
			}
		}
		m.phase = PhaseCleanup
		m.syncPresentation(runstate.PhaseCleanup)
		m.markMainScrollJump()

	case events.EventImplementationReviewCompleted:
		outcome := "clean"
		if !e.Clean {
			outcome = "findings"
		}
		m.logger.AddLog(fmt.Sprintf("Cleanup completed (iteration %d, %s)", e.Iteration, outcome))
		if e.Clean {
			m.activity = session.RunActivity{Kind: session.ActivityCleanup}
		}
		m.phase = PhaseCleanup
		m.syncPresentation(runstate.PhaseCleanup)
		m.markMainScrollJump()

	case events.EventRecoveryStarted:
		_, storyID, storyTitle := m.activeStoryForActivity()
		m.activity = session.RunActivity{
			Kind:        session.ActivityRecovery,
			StoryID:     storyID,
			StoryTitle:  storyTitle,
			Attempt:     e.Attempt,
			MaxAttempts: e.Max,
		}
		if m.phase == PhaseCleanup {
			m.syncPresentation(runstate.PhaseCleanup)
		} else {
			m.phase = PhaseImplementation
			m.syncPresentation(runstate.PhaseImplement)
		}
		m.logger.AddLog(fmt.Sprintf("Recovery started (%s, attempt %d/%d)", e.Reason, e.Attempt, e.Max))
		m.markMainScrollJump()

	case events.EventRecoveryCompleted:
		outcome := "failed"
		if e.Success {
			outcome = "succeeded"
		}
		m.logger.AddLog(fmt.Sprintf("Recovery %s (%s, attempt %d)", outcome, e.Reason, e.Attempt))
		if e.Success && m.phase == PhaseCleanup {
			m.activity = session.RunActivity{Kind: session.ActivityReview}
			m.syncPresentation(runstate.PhaseCleanup)
		} else {
			m.syncPresentation(runstate.PhaseImplement)
		}
		m.markMainScrollJump()

	case events.EventCleanupStarted:
		m.activity = session.RunActivity{Kind: session.ActivityCleanup}
		m.phase = PhaseCleanup
		m.logger.AddLog("Running post-implementation cleanup...")
		m.syncPresentation(runstate.PhaseCleanup)
		m.markMainScrollJump()

	case events.EventCleanupCompleted:
		m.logger.AddLog("Cleanup complete!")
		m.markMainScrollJump()

	case events.EventOutput:
		if !e.Verbose || m.verbose {
			m.logger.AddOutputLine(runner.OutputLine{Text: e.Text, IsErr: e.IsErr, Append: e.Append})
		}

	case events.EventError:
		m.logger.AddLog(fmt.Sprintf("Error: %v", e.Err))
		m.retryImplementation = m.phase == PhaseImplementation
		m.revisingPRD = false
		m.err = e.Err
		m.phase = PhaseFailed
		m.markMainScrollJump()

	case events.EventCompleted:
		m.activity = session.RunActivity{}
		m.retryImplementation = false
		m.err = nil
		m.phase = PhaseCompleted
		m.logger.AddLog("All stories completed!")
		m.markMainScrollJump()
	}

	return nil
}
