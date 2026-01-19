package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ralph/internal"
	"ralph/internal/config"
	"ralph/internal/prd"
	"ralph/internal/runner"
	"ralph/internal/story"
)

type Phase int

const (
	PhaseInit Phase = iota
	PhasePRDGeneration
	PhaseImplementation
	PhaseCompleted
	PhaseFailed
)

func (p Phase) String() string {
	switch p {
	case PhaseInit:
		return "Initializing"
	case PhasePRDGeneration:
		return "Phase 1: PRD Generation"
	case PhaseImplementation:
		return "Phase 2: Implementation"
	case PhaseCompleted:
		return "Completed"
	case PhaseFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

type Model struct {
	cfg     *config.Config
	prompt  string
	dryRun  bool
	resume  bool
	verbose bool

	phase        Phase
	prd          *prd.PRD
	currentStory *prd.Story
	iteration    int
	err          error
	quitting     bool
	width        int
	height       int

	spinner  spinner.Model
	progress progress.Model

	logger           *Logger
	operationManager *OperationManager
	phaseHandler     PhaseHandler
}

func NewModel(cfg *config.Config, prompt string, dryRun, resume, verbose bool) *Model {
	s := spinner.New()
	s.Spinner = spinner.Jump
	s.Style = lipgloss.NewStyle().Foreground(primaryColor)

	p := progress.New(
		progress.WithGradient("#8B5CF6", "#34D399"),
		progress.WithWidth(40),
		progress.WithSolidFill("#374151"),
	)

	logger := NewLogger(verbose)
	operationManager := NewOperationManager(cfg, prd.NewGenerator(cfg), story.NewImplementer(cfg))

	return &Model{
		cfg:              cfg,
		prompt:           prompt,
		dryRun:           dryRun,
		resume:           resume,
		verbose:          verbose,
		phase:            PhaseInit,
		spinner:          s,
		progress:         p,
		logger:           logger,
		operationManager: operationManager,
		phaseHandler:     NewInitPhaseHandler(),
	}
}

// SetGenerator sets the PRD generator (for testing)
func (m *Model) SetGenerator(g internal.PRDGenerator) {
	m.operationManager = NewOperationManager(m.cfg, g, m.operationManager.implementer)
}

// SetImplementer sets the story implementer (for testing)
func (m *Model) SetImplementer(i internal.StoryImplementer) {
	m.operationManager = NewOperationManager(m.cfg, m.operationManager.generator, i)
}

type (
	outputMsg        runner.OutputLine
	prdGeneratedMsg  struct{ prd *prd.PRD }
	prdErrorMsg      struct{ err error }
	storyStartMsg    struct{ story *prd.Story }
	storyCompleteMsg struct{ success bool }
	storyErrorMsg    struct{ err error }
	phaseChangeMsg   Phase
	tickMsg          time.Time
)

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.operationManager.StartOperation(),
		m.operationManager.ListenForOutput(),
		tea.WindowSize(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quitting = true
			m.operationManager.Cancel()
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.logger.SetSize(msg.Width, msg.Height)
		m.progress.Width = min(40, max(10, msg.Width-20))

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case outputMsg:
		m.logger.AddOutputLine(runner.OutputLine(msg))
		cmds = append(cmds, m.operationManager.ListenForOutput())

	case phaseChangeMsg:
		m.phase = Phase(msg)
		m.phaseHandler = m.getPhaseHandlerForPhase(m.phase)
		// When entering PRD generation phase, start the actual operation
		if m.phase == PhasePRDGeneration {
			cmds = append(cmds, func() tea.Msg {
				return m.operationManager.RunPRDOperation(m.resume, m.prompt)
			})
		}

	case prdGeneratedMsg:
		m.prd = msg.prd
		m.logger.AddLog(fmt.Sprintf("PRD generated: %s (%d stories)", m.prd.ProjectName, len(m.prd.Stories)))

		if m.dryRun {
			m.phase = PhaseCompleted
			m.logger.AddLog("Dry run complete - PRD saved to " + m.cfg.PRDFile)
		} else {
			m.phase = PhaseImplementation
			cmds = append(cmds, m.operationManager.SetupBranchAndStart(m.prd.BranchName))
		}

	case prdErrorMsg:
		m.err = msg.err
		m.phase = PhaseFailed
		m.logger.AddLog(fmt.Sprintf("Error: %v", msg.err))

	case storyStartMsg:
		m.currentStory = msg.story
		m.iteration++
		m.logger.AddLog(fmt.Sprintf("Starting story: %s (attempt %d/%d)",
			msg.story.Title, msg.story.RetryCount+1, m.cfg.RetryAttempts))
		// Re-register the output listener for the new story's output
		cmds = append(cmds, m.operationManager.ListenForOutput())

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
		cmds = append(cmds, m.operationManager.ContinueImplementation(m.prd, m.iteration))

	case storyErrorMsg:
		m.logger.AddLog(fmt.Sprintf("Error: %v", msg.err))
		m.currentStory.RetryCount++
		cmds = append(cmds, m.operationManager.ContinueImplementation(m.prd, m.iteration))

	default:
		// Delegate to phase handler for any remaining messages
		if m.phaseHandler != nil {
			model, cmd := m.phaseHandler.HandleUpdate(msg, m)
			if model_, ok := model.(*Model); ok {
				m = model_
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	var cmd tea.Cmd
	m.logger.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) ExitCode() int {
	switch m.phase {
	case PhaseCompleted:
		return 0
	case PhaseFailed:
		if m.prd != nil && m.prd.CompletedCount() > 0 {
			return 2
		}
		return 1
	default:
		return 1
	}
}

func (m *Model) getPhaseHandlerForPhase(phase Phase) PhaseHandler {
	switch phase {
	case PhaseInit:
		return NewInitPhaseHandler()
	case PhasePRDGeneration:
		return NewPRDGenerationPhaseHandler()
	case PhaseImplementation:
		return NewImplementationPhaseHandler()
	case PhaseCompleted:
		return NewCompletedPhaseHandler()
	case PhaseFailed:
		return NewFailedPhaseHandler()
	default:
		return NewInitPhaseHandler()
	}
}
