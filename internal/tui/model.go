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
	"ralph/internal/git"
	"ralph/internal/prd"
	"ralph/internal/runner"
	"ralph/internal/story"
)

// Phase represents the current phase of execution
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

// Model is the main Bubble Tea model
type Model struct {
	// Configuration
	cfg     *config.Config
	prompt  string
	dryRun  bool
	resume  bool
	workDir string

	// State
	phase        Phase
	prd          *prd.PRD
	currentStory *prd.Story
	iteration    int
	err          error
	quitting     bool
	width        int
	height       int

	// Components
	spinner  spinner.Model
	progress progress.Model
	logView  viewport.Model
	logs     []string
	maxLogs  int

	// Background operation
	ctx        context.Context
	cancelFunc context.CancelFunc
	outputCh   chan runner.OutputLine
}

// NewModel creates a new TUI model
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
		cfg:        cfg,
		prompt:     prompt,
		dryRun:     dryRun,
		resume:     resume,
		phase:      PhaseInit,
		spinner:    s,
		progress:   p,
		logView:    v,
		logs:       make([]string, 0),
		maxLogs:    100,
		ctx:        ctx,
		cancelFunc: cancel,
		outputCh:   make(chan runner.OutputLine, 100),
	}
}

// Messages

type outputMsg runner.OutputLine
type prdGeneratedMsg struct{ prd *prd.PRD }
type prdErrorMsg struct{ err error }
type storyStartMsg struct{ story *prd.Story }
type storyCompleteMsg struct{ success bool }
type storyErrorMsg struct{ err error }
type phaseChangeMsg Phase
type tickMsg time.Time

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.startOperation(),
		m.listenForOutput(),
		tea.WindowSize(),
	)
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
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
		m.addLog(fmt.Sprintf("Starting story: %s (attempt %d/%d)", msg.story.Title, msg.story.RetryCount+1, m.cfg.RetryAttempts))

	case storyCompleteMsg:
		if msg.success {
			m.currentStory.Passes = true
			m.addLog(fmt.Sprintf("Story completed: %s", m.currentStory.Title))
		} else {
			m.currentStory.RetryCount++
			m.addLog(fmt.Sprintf("Story failed: %s (retry %d/%d)", m.currentStory.Title, m.currentStory.RetryCount, m.cfg.RetryAttempts))
		}

		// Save state
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
	}

	// Update viewport
	var cmd tea.Cmd
	m.logView, cmd = m.logView.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m *Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Phase indicator
	b.WriteString(m.renderPhase())
	b.WriteString("\n")

	// Main content based on phase
	switch m.phase {
	case PhaseInit, PhasePRDGeneration:
		b.WriteString(m.renderGenerating())
	case PhaseImplementation:
		b.WriteString(m.renderImplementation())
	case PhaseCompleted:
		b.WriteString(m.renderCompleted())
	case PhaseFailed:
		b.WriteString(m.renderFailed())
	}

	// Log viewport
	b.WriteString("\n")
	b.WriteString(m.renderLogs())

	// Help
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press q to quit"))

	return b.String()
}

func (m *Model) renderHeader() string {
	title := "🤖 RALPH - Autonomous Software Development Agent"
	return headerStyle.Render(title)
}

func (m *Model) renderPhase() string {
	icon := m.spinner.View()
	if m.phase == PhaseCompleted {
		icon = "✓"
	} else if m.phase == PhaseFailed {
		icon = "✗"
	}
	return phaseStyle.Render(fmt.Sprintf("%s %s", icon, m.phase.String()))
}

func (m *Model) renderGenerating() string {
	var b strings.Builder

	b.WriteString(boxStyle.Render(fmt.Sprintf(
		"Prompt: %s\n\nGenerating PRD from your requirements...",
		truncate(m.prompt, 60),
	)))

	return b.String()
}

func (m *Model) renderImplementation() string {
	if m.prd == nil {
		return ""
	}

	var b strings.Builder

	// Project info
	b.WriteString(fmt.Sprintf("📁 Project: %s\n", m.prd.ProjectName))
	if m.prd.BranchName != "" {
		b.WriteString(fmt.Sprintf("🌿 Branch: %s\n", m.prd.BranchName))
	}
	b.WriteString("\n")

	// Progress bar
	completed := m.prd.CompletedCount()
	total := len(m.prd.Stories)
	percent := float64(completed) / float64(total)

	b.WriteString(fmt.Sprintf("Progress: %d/%d stories ", completed, total))
	b.WriteString(m.progress.ViewAs(percent))
	b.WriteString("\n\n")

	// Story list
	b.WriteString("Stories:\n")
	for _, s := range m.prd.Stories {
		isCurrentStory := m.currentStory != nil && s.ID == m.currentStory.ID
		icon := getStatusIcon(s.Passes, isCurrentStory, s.RetryCount, m.cfg.RetryAttempts)
		status := getStatusText(s.Passes, isCurrentStory, s.RetryCount, m.cfg.RetryAttempts)

		line := fmt.Sprintf("%s %s [%s]", icon, s.Title, status)
		if isCurrentStory {
			b.WriteString(selectedStoryStyle.Render(line))
		} else {
			b.WriteString(storyItemStyle.Render(line))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m *Model) renderCompleted() string {
	var b strings.Builder

	if m.dryRun {
		b.WriteString(successStyle.Render("✓ Dry run completed!\n\n"))
		b.WriteString(fmt.Sprintf("PRD saved to: %s\n", m.cfg.PRDFile))
		b.WriteString("Run without --dry-run to implement, or use --resume.\n")
	} else if m.prd != nil {
		b.WriteString(successStyle.Render("✓ All stories completed!\n\n"))
		b.WriteString(fmt.Sprintf("📁 Project: %s\n", m.prd.ProjectName))
		b.WriteString(fmt.Sprintf("📊 Stories: %d completed\n", len(m.prd.Stories)))
		b.WriteString(fmt.Sprintf("📝 Iterations: %d\n", m.iteration))
	}

	return boxStyle.Render(b.String())
}

func (m *Model) renderFailed() string {
	var b strings.Builder

	b.WriteString(errorStyle.Render("✗ Implementation failed\n\n"))

	if m.err != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n", m.err))
	}

	if m.prd != nil {
		failed := m.prd.FailedStories(m.cfg.RetryAttempts)
		if len(failed) > 0 {
			b.WriteString(fmt.Sprintf("\nFailed stories (%d):\n", len(failed)))
			for _, s := range failed {
				b.WriteString(fmt.Sprintf("  • %s (%d attempts)\n", s.Title, s.RetryCount))
			}
		}
		b.WriteString("\nRun with --resume to retry after fixing issues.\n")
	}

	return boxStyle.Render(b.String())
}

func (m *Model) renderLogs() string {
	if len(m.logs) == 0 {
		return logBoxStyle.Render("Waiting for output...")
	}

	// Show last N lines
	startIdx := 0
	if len(m.logs) > 8 {
		startIdx = len(m.logs) - 8
	}

	var lines []string
	for i := startIdx; i < len(m.logs); i++ {
		lines = append(lines, logLineStyle.Render(truncate(m.logs[i], m.width-6)))
	}

	return logBoxStyle.Render(strings.Join(lines, "\n"))
}

func (m *Model) addLog(line string) {
	m.logs = append(m.logs, line)
	if len(m.logs) > m.maxLogs {
		m.logs = m.logs[1:]
	}
	m.logView.SetContent(strings.Join(m.logs, "\n"))
	m.logView.GotoBottom()
}

// Commands

func (m *Model) startOperation() tea.Cmd {
	return func() tea.Msg {
		if m.resume {
			return m.loadAndResume()
		}
		return m.generatePRD()
	}
}

func (m *Model) loadAndResume() tea.Msg {
	loadedPRD, err := prd.Load(m.cfg)
	if err != nil {
		return prdErrorMsg{err: err}
	}
	return prdGeneratedMsg{prd: loadedPRD}
}

func (m *Model) generatePRD() tea.Msg {
	gen := prd.NewGenerator(m.cfg)

	generatedPRD, err := gen.Generate(m.ctx, m.prompt, m.outputCh)
	if err != nil {
		return prdErrorMsg{err: err}
	}

	// Save the PRD
	if err := prd.Save(m.cfg, generatedPRD); err != nil {
		return prdErrorMsg{err: fmt.Errorf("failed to save PRD: %w", err)}
	}

	return prdGeneratedMsg{prd: generatedPRD}
}

func (m *Model) setupBranchAndStart() tea.Cmd {
	return func() tea.Msg {
		if m.prd.BranchName != "" {
			gitMgr := git.New()
			if err := gitMgr.CreateBranch(m.prd.BranchName); err != nil {
				m.addLog(fmt.Sprintf("Warning: failed to create branch: %v", err))
			}
		}
		return m.startNextStory()
	}
}

func (m *Model) continueImplementation() tea.Cmd {
	return func() tea.Msg {
		// Check if all completed
		if m.prd.AllCompleted() {
			// Cleanup PRD file on success
			prd.Delete(m.cfg)
			return phaseChangeMsg(PhaseCompleted)
		}

		// Check for next story
		next := m.prd.NextPendingStory(m.cfg.RetryAttempts)
		if next == nil {
			// All remaining stories have failed
			return phaseChangeMsg(PhaseFailed)
		}

		// Check max iterations
		if m.iteration >= m.cfg.MaxIterations {
			return phaseChangeMsg(PhaseFailed)
		}

		return m.startNextStory()
	}
}

func (m *Model) startNextStory() tea.Msg {
	next := m.prd.NextPendingStory(m.cfg.RetryAttempts)
	if next == nil {
		if m.prd.AllCompleted() {
			return phaseChangeMsg(PhaseCompleted)
		}
		return phaseChangeMsg(PhaseFailed)
	}

	// Start implementing the story
	go func() {
		impl := story.NewImplementer(m.cfg)
		success, err := impl.Implement(m.ctx, next, m.iteration+1, m.prd, m.outputCh)

		if err != nil {
			m.outputCh <- runner.OutputLine{Text: fmt.Sprintf("Error: %v", err), IsErr: true}
		}

		// Signal completion through output channel
		if success {
			m.outputCh <- runner.OutputLine{Text: "STORY_COMPLETE:success"}
		} else {
			m.outputCh <- runner.OutputLine{Text: "STORY_COMPLETE:failure"}
		}
	}()

	return storyStartMsg{story: next}
}

func (m *Model) listenForOutput() tea.Cmd {
	return func() tea.Msg {
		select {
		case <-m.ctx.Done():
			return nil
		case line, ok := <-m.outputCh:
			if !ok {
				return nil
			}
			// Handle special completion markers
			if strings.HasPrefix(line.Text, "STORY_COMPLETE:") {
				success := strings.HasSuffix(line.Text, "success")
				return storyCompleteMsg{success: success}
			}
			return outputMsg(line)
		}
	}
}

// ExitCode returns the appropriate exit code based on final state
func (m *Model) ExitCode() int {
	switch m.phase {
	case PhaseCompleted:
		return 0
	case PhaseFailed:
		if m.prd != nil && m.prd.CompletedCount() > 0 {
			return 2 // Partial success
		}
		return 1
	default:
		return 1
	}
}

// Helpers

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max < 4 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
