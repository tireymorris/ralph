package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	cfg    *config.Config
	prompt string
	dryRun bool
	resume bool

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
	logView  viewport.Model
	logs     []string
	maxLogs  int

	ctx         context.Context
	cancelFunc  context.CancelFunc
	outputCh    chan runner.OutputLine
	generator   PRDGenerator
	implementer StoryImplementer
}

func NewModel(cfg *config.Config, prompt string, dryRun, resume bool) *Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(primaryColor)

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	v := viewport.New(80, 10)
	v.Style = logBoxStyle

	ctx, cancel := context.WithCancel(context.Background())

	return &Model{
		cfg:         cfg,
		prompt:      prompt,
		dryRun:      dryRun,
		resume:      resume,
		phase:       PhaseInit,
		spinner:     s,
		progress:    p,
		logView:     v,
		logs:        make([]string, 0),
		maxLogs:     500,
		ctx:         ctx,
		cancelFunc:  cancel,
		outputCh:    make(chan runner.OutputLine, 100),
		generator:   prd.NewGenerator(cfg),
		implementer: story.NewImplementer(cfg),
	}
}

func (m *Model) SetGenerator(g PRDGenerator) {
	m.generator = g
}

func (m *Model) SetImplementer(i StoryImplementer) {
	m.implementer = i
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
		m.startOperation(),
		m.listenForOutput(),
		tea.WindowSize(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quitting = true
			if m.cancelFunc != nil {
				m.cancelFunc()
			}
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.logView.Width = msg.Width - 4
		m.logView.Height = min(10, msg.Height/3)
		m.progress.Width = min(40, msg.Width-20)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case outputMsg:
		m.addLog(msg.Text)
		cmds = append(cmds, m.listenForOutput())

	case prdGeneratedMsg:
		m.prd = msg.prd
		m.addLog(fmt.Sprintf("PRD generated: %s (%d stories)", m.prd.ProjectName, len(m.prd.Stories)))

		if m.dryRun {
			m.phase = PhaseCompleted
			m.addLog("Dry run complete - PRD saved to " + m.cfg.PRDFile)
		} else {
			m.phase = PhaseImplementation
			cmds = append(cmds, m.setupBranchAndStart())
		}

	case prdErrorMsg:
		m.err = msg.err
		m.phase = PhaseFailed
		m.addLog(fmt.Sprintf("Error: %v", msg.err))

	case storyStartMsg:
		m.currentStory = msg.story
		m.iteration++
		m.addLog(fmt.Sprintf("Starting story: %s (attempt %d/%d)",
			msg.story.Title, msg.story.RetryCount+1, m.cfg.RetryAttempts))

	case storyCompleteMsg:
		if msg.success {
			m.currentStory.Passes = true
			m.addLog(fmt.Sprintf("Story completed: %s", m.currentStory.Title))
		} else {
			m.currentStory.RetryCount++
			m.addLog(fmt.Sprintf("Story failed: %s (retry %d/%d)",
				m.currentStory.Title, m.currentStory.RetryCount, m.cfg.RetryAttempts))
		}

		if err := prd.Save(m.cfg, m.prd); err != nil {
			m.addLog(fmt.Sprintf("Warning: failed to save state: %v", err))
		}
		cmds = append(cmds, m.continueImplementation())

	case storyErrorMsg:
		m.addLog(fmt.Sprintf("Error: %v", msg.err))
		m.currentStory.RetryCount++
		cmds = append(cmds, m.continueImplementation())

	case phaseChangeMsg:
		m.phase = Phase(msg)
		// When entering PRD generation phase, start the actual operation
		if m.phase == PhasePRDGeneration {
			cmds = append(cmds, m.runPRDOperation())
		}
	}

	var cmd tea.Cmd
	m.logView, cmd = m.logView.Update(msg)
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

func (m *Model) addLog(line string) {
	m.logs = append(m.logs, line)
	if len(m.logs) > m.maxLogs {
		m.logs = m.logs[1:]
	}
	m.logView.SetContent(strings.Join(m.logs, "\n"))
	m.logView.GotoBottom()
}
