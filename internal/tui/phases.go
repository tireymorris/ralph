package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"ralph/internal/prd"
)

// PhaseHandler defines the interface for handling different phases
type PhaseHandler interface {
	HandleUpdate(msg tea.Msg, m *Model) (tea.Model, tea.Cmd)
	GetPhase() Phase
}

// BasePhaseHandler provides common functionality for all phase handlers
type BasePhaseHandler struct {
	phase Phase
}

func (h *BasePhaseHandler) GetPhase() Phase {
	return h.phase
}

// InitPhaseHandler handles the initialization phase
type InitPhaseHandler struct {
	BasePhaseHandler
}

func NewInitPhaseHandler() *InitPhaseHandler {
	return &InitPhaseHandler{
		BasePhaseHandler: BasePhaseHandler{phase: PhaseInit},
	}
}

func (h *InitPhaseHandler) HandleUpdate(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	// In init phase, we mostly delegate to the main update logic
	// but phase-specific handling can be added here
	return m, nil
}

// PRDGenerationPhaseHandler handles the PRD generation phase
type PRDGenerationPhaseHandler struct {
	BasePhaseHandler
}

func NewPRDGenerationPhaseHandler() *PRDGenerationPhaseHandler {
	return &PRDGenerationPhaseHandler{
		BasePhaseHandler: BasePhaseHandler{phase: PhasePRDGeneration},
	}
}

func (h *PRDGenerationPhaseHandler) HandleUpdate(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case prdGeneratedMsg:
		m.prd = msg.prd
		m.logger.AddLog(fmt.Sprintf("PRD generated: %s (%d stories)", m.prd.ProjectName, len(m.prd.Stories)))

		if m.dryRun {
			m.phase = PhaseCompleted
			m.logger.AddLog("Dry run complete - PRD saved to " + m.cfg.PRDFile)
			return m, nil
		} else {
			m.phase = PhaseImplementation
			return m, m.operationManager.SetupBranchAndStart(m.prd.BranchName)
		}

	case prdErrorMsg:
		m.err = msg.err
		m.phase = PhaseFailed
		m.logger.AddLog(fmt.Sprintf("Error: %v", msg.err))
		return m, nil

	default:
		return m, nil
	}
}

// ImplementationPhaseHandler handles the implementation phase
type ImplementationPhaseHandler struct {
	BasePhaseHandler
}

func NewImplementationPhaseHandler() *ImplementationPhaseHandler {
	return &ImplementationPhaseHandler{
		BasePhaseHandler: BasePhaseHandler{phase: PhaseImplementation},
	}
}

func (h *ImplementationPhaseHandler) HandleUpdate(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case storyStartMsg:
		m.currentStory = msg.story
		m.iteration++
		m.logger.AddLog(fmt.Sprintf("Starting story: %s (attempt %d/%d)",
			msg.story.Title, msg.story.RetryCount+1, m.cfg.RetryAttempts))
		// Re-register the output listener for the new story's output
		return m, m.operationManager.ListenForOutput()

	case storyCompleteMsg:
		if msg.success {
			m.currentStory.Passes = true
			m.logger.AddLog(fmt.Sprintf("Story completed: %s", m.currentStory.Title))
		} else {
			m.currentStory.RetryCount++
			m.logger.AddLog(fmt.Sprintf("Story failed: %s (retry %d/%d)",
				m.currentStory.Title, m.currentStory.RetryCount, m.cfg.RetryAttempts))
		}

		if err := prd.Save(m.cfg, m.prd); err != nil {
			m.logger.AddLog(fmt.Sprintf("Warning: failed to save state: %v", err))
		}
		return m, m.operationManager.ContinueImplementation(m.prd, m.iteration)

	case storyErrorMsg:
		m.logger.AddLog(fmt.Sprintf("Error: %v", msg.err))
		m.currentStory.RetryCount++
		return m, m.operationManager.ContinueImplementation(m.prd, m.iteration)

	default:
		return m, nil
	}
}

// CompletedPhaseHandler handles the completed phase
type CompletedPhaseHandler struct {
	BasePhaseHandler
}

func NewCompletedPhaseHandler() *CompletedPhaseHandler {
	return &CompletedPhaseHandler{
		BasePhaseHandler: BasePhaseHandler{phase: PhaseCompleted},
	}
}

func (h *CompletedPhaseHandler) HandleUpdate(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	// In completed phase, only handle quit commands
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quitting = true
			m.operationManager.Cancel()
			return m, tea.Quit
		}
	default:
		return m, nil
	}
	return m, nil
}

// FailedPhaseHandler handles the failed phase
type FailedPhaseHandler struct {
	BasePhaseHandler
}

func NewFailedPhaseHandler() *FailedPhaseHandler {
	return &FailedPhaseHandler{
		BasePhaseHandler: BasePhaseHandler{phase: PhaseFailed},
	}
}

func (h *FailedPhaseHandler) HandleUpdate(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	// In failed phase, only handle quit commands
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quitting = true
			m.operationManager.Cancel()
			return m, tea.Quit
		}
	default:
		return m, nil
	}
	return m, nil
}
