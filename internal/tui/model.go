package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ralph/internal/config"
	"ralph/internal/prd"
	"ralph/internal/runner"
	"ralph/internal/workflow"
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
}

func NewModel(cfg *config.Config, prompt string, dryRun, resume, verbose bool) *Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(accentColor)

	p := progress.New(
		progress.WithGradient("#A855F7", "#10B981"),
		progress.WithWidth(40),
		progress.WithSolidFill("#4B5563"),
	)

	logger := NewLogger(verbose)
	operationManager := NewOperationManager(cfg)

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
	}
}

type (
	prdGeneratedMsg struct{ prd *prd.PRD }
	prdErrorMsg     struct{ err error }
	phaseChangeMsg  Phase
)

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.operationManager.StartOperation(),
		m.operationManager.ListenForEvents(),
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

	case phaseChangeMsg:
		m.phase = Phase(msg)
		if m.phase == PhasePRDGeneration {
			cmds = append(cmds, func() tea.Msg {
				return m.operationManager.RunPRDOperation(m.resume, m.prompt)
			})
		}

	case prdGeneratedMsg:
		m.prd = msg.prd
		m.logger.AddLog(fmt.Sprintf("PRD: %s (%d stories)", m.prd.ProjectName, len(m.prd.Stories)))

		if m.dryRun {
			m.phase = PhaseCompleted
			m.logger.AddLog("Dry run complete - PRD saved to " + m.cfg.PRDFile)
		} else {
			m.phase = PhaseImplementation
			cmds = append(cmds, m.operationManager.StartImplementation(m.prd))
			cmds = append(cmds, m.operationManager.ListenForEvents())
		}

	case prdErrorMsg:
		m.err = msg.err
		m.phase = PhaseFailed
		m.logger.AddLog(fmt.Sprintf("Error: %v", msg.err))

	case workflowEventMsg:
		cmds = append(cmds, m.handleWorkflowEvent(msg.event))
		cmds = append(cmds, m.operationManager.ListenForEvents())
	}

	_, logCmd := m.logger.Update(msg)
	if cmd, ok := logCmd.(tea.Cmd); ok && cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleWorkflowEvent(event workflow.Event) tea.Cmd {
	switch e := event.(type) {
	case workflow.EventPRDGenerating:
		m.logger.AddLog("Generating PRD...")

	case workflow.EventPRDGenerated:
		m.prd = e.PRD
		m.logger.AddLog(fmt.Sprintf("PRD generated: %s (%d stories)", e.PRD.ProjectName, len(e.PRD.Stories)))

	case workflow.EventPRDLoaded:
		m.prd = e.PRD
		m.logger.AddLog(fmt.Sprintf("Loaded PRD: %s (%d/%d completed)",
			e.PRD.ProjectName, e.PRD.CompletedCount(), len(e.PRD.Stories)))

	case workflow.EventStoryStarted:
		m.currentStory = e.Story
		m.iteration = e.Iteration
		m.logger.AddLog(fmt.Sprintf("Starting: %s (attempt %d/%d)",
			e.Story.Title, e.Story.RetryCount+1, m.cfg.RetryAttempts))

	case workflow.EventStoryCompleted:
		if e.Success {
			m.logger.AddLog(fmt.Sprintf("Completed: %s", e.Story.Title))
			if m.prd != nil {
				if s := m.prd.GetStory(e.Story.ID); s != nil {
					s.Passes = true
				}
			}
		} else {
			m.logger.AddLog(fmt.Sprintf("Failed: %s", e.Story.Title))
		}

	case workflow.EventOutput:
		if !e.Verbose || m.verbose {
			m.logger.AddOutputLine(runner.OutputLine{Text: e.Text, IsErr: e.IsErr})
		}

	case workflow.EventError:
		m.logger.AddLog(fmt.Sprintf("Error: %v", e.Err))

	case workflow.EventCompleted:
		m.phase = PhaseCompleted
		m.logger.AddLog("All stories completed!")

	case workflow.EventFailed:
		m.phase = PhaseFailed
		if len(e.FailedStories) > 0 {
			m.logger.AddLog(fmt.Sprintf("Failed: %d stories exceeded retry limit", len(e.FailedStories)))
		}
	}

	return nil
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
