package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ralph/internal/config"
	"ralph/internal/prd"
	"ralph/internal/prompt"
	"ralph/internal/runner"
	"ralph/internal/workflow"
)

type Phase int

const (
	PhaseInit Phase = iota
	PhaseClarifying
	PhasePRDGeneration
	PhasePRDReview
	PhaseImplementation
	PhaseCompleted
	PhaseFailed
)

func (p Phase) String() string {
	switch p {
	case PhaseInit:
		return "Initializing"
	case PhaseClarifying:
		return "Clarifying Questions"
	case PhasePRDGeneration:
		return "Phase 1: PRD Generation"
	case PhasePRDReview:
		return "PRD Review"
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

	mainPane      viewport.Model
	scrollPane    scrollFocus
	snapMainToTop bool

	logHeightBias int

	layoutSigW, layoutSigH int
	layoutSigLogCount      int
	layoutSigBias          int

	// Clarifying questions state
	clarifyQuestions []string
	clarifyInputs    []textinput.Model
	clarifyFocused   int
	clarifyAnswersCh chan<- []prompt.QuestionAnswer

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

	mv := viewport.New(80, 12)
	mv.Style = lipgloss.NewStyle()
	mv.MouseWheelEnabled = true

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
		mainPane:         mv,
		scrollPane:       focusMain,
		logger:           logger,
		operationManager: operationManager,
	}
}

func (m *Model) applyLayout(width, height int) {
	lc := m.logger.LogCount()
	bias := m.logHeightBias
	mainH, logH := computePaneHeights(height, lc, bias)
	if width == m.layoutSigW && height == m.layoutSigH && lc == m.layoutSigLogCount && bias == m.layoutSigBias {
		return
	}
	m.layoutSigW = width
	m.layoutSigH = height
	m.layoutSigLogCount = lc
	m.layoutSigBias = bias
	m.width = width
	m.height = height
	m.logger.SetSize(width, logH)
	m.mainPane.Width = max(20, width-4)
	m.mainPane.Height = max(4, mainH)
	m.progress.Width = min(40, max(10, width-20))
}

func (m *Model) markMainScrollJump() {
	m.snapMainToTop = true
	m.scrollPane = focusMain
}

func (m *Model) splitScrollMsg(msg tea.Msg) (tea.Msg, tea.Msg) {
	if !isScrollNavMsg(msg) {
		return msg, msg
	}
	if m.scrollPane == focusMain {
		return msg, noopScrollMsg
	}
	return noopScrollMsg, msg
}

func (m *Model) syncAfterClarifySubmit() {
	m.scrollPane = focusMain
	if m.width > 0 && m.height > 0 {
		m.applyLayout(m.width, m.height)
	}
	m.rebuildMainScrollContent()
	m.mainPane.GotoTop()
}

type (
	prdGeneratedMsg struct{ prd *prd.PRD }
	prdErrorMsg     struct{ err error }
	phaseChangeMsg  Phase

	clarifyQuestionsMsg struct {
		questions []string
		answersCh chan<- []prompt.QuestionAnswer
	}
)

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.operationManager.ListenForEvents(),
		tea.WindowSize(),
		m.operationManager.StartFullOperation(m.resume, m.prompt),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var needsMainRebuild bool

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// In clarifying phase, route key events to the active input
		if m.phase == PhaseClarifying && len(m.clarifyInputs) > 0 {
			switch msg.String() {
			case "ctrl+c":
				m.quitting = true
				m.operationManager.Cancel()
				return m, tea.Quit
			case "esc":
				// Skip all questions, proceed with PRD generation as-is
				submitCmds := m.submitClarifyingAnswers(nil)
				m.syncAfterClarifySubmit()
				return m, tea.Batch(submitCmds...)
			case "tab", "down":
				m.clarifyInputs[m.clarifyFocused].Blur()
				m.clarifyFocused = (m.clarifyFocused + 1) % len(m.clarifyInputs)
				m.clarifyInputs[m.clarifyFocused].Focus()
				cmds = append(cmds, textinput.Blink)
			case "shift+tab", "up":
				m.clarifyInputs[m.clarifyFocused].Blur()
				m.clarifyFocused = (m.clarifyFocused - 1 + len(m.clarifyInputs)) % len(m.clarifyInputs)
				m.clarifyInputs[m.clarifyFocused].Focus()
				cmds = append(cmds, textinput.Blink)
			case "enter":
				if m.clarifyFocused < len(m.clarifyInputs)-1 {
					// Move to next input
					m.clarifyInputs[m.clarifyFocused].Blur()
					m.clarifyFocused++
					m.clarifyInputs[m.clarifyFocused].Focus()
					cmds = append(cmds, textinput.Blink)
				} else {
					// Last field — submit
					cmds = append(cmds, m.submitClarifyingAnswers(m.buildAnswers())...)
					m.syncAfterClarifySubmit()
				}
			default:
				var cmd tea.Cmd
				m.clarifyInputs[m.clarifyFocused], cmd = m.clarifyInputs[m.clarifyFocused].Update(msg)
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quitting = true
			m.operationManager.Cancel()
			return m, tea.Quit
		}

		if m.mainScrollEnabled() && msg.String() == "tab" {
			if m.scrollPane == focusMain {
				m.scrollPane = focusLogs
			} else {
				m.scrollPane = focusMain
			}
		}

		if m.mainScrollEnabled() {
			switch msg.String() {
			case "[":
				m.logHeightBias--
				if m.logHeightBias < -logBiasMaxLines {
					m.logHeightBias = -logBiasMaxLines
				}
			case "]":
				m.logHeightBias++
				if m.logHeightBias > logBiasMaxLines {
					m.logHeightBias = logBiasMaxLines
				}
			}
		}

		// In PRD review phase, Enter proceeds to implementation
		if m.phase == PhasePRDReview && msg.String() == "enter" {
			if m.prd != nil {
				m.phase = PhaseImplementation
				m.scrollPane = focusMain
				if m.width > 0 && m.height > 0 {
					m.applyLayout(m.width, m.height)
				}
				m.rebuildMainScrollContent()
				m.mainPane.GotoTop()
				return m, m.operationManager.StartImplementation(m.prd)
			}
		}

	case tea.WindowSizeMsg:
		needsMainRebuild = true
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		needsMainRebuild = true
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case phaseChangeMsg:
		m.phase = Phase(msg)
		needsMainRebuild = true

	case clarifyQuestionsMsg:
		m.phase = PhaseClarifying
		m.clarifyQuestions = msg.questions
		m.clarifyAnswersCh = msg.answersCh
		m.clarifyFocused = 0
		m.clarifyInputs = make([]textinput.Model, len(msg.questions))
		for i := range m.clarifyInputs {
			ti := textinput.New()
			ti.Placeholder = "Your answer (press Enter to continue, Esc to skip all)"
			ti.CharLimit = 500
			if i == 0 {
				ti.Focus()
			}
			m.clarifyInputs[i] = ti
		}
		cmds = append(cmds, textinput.Blink)
		cmds = append(cmds, m.operationManager.ListenForEvents())

	case workflowEventMsg:
		needsMainRebuild = true
		cmds = append(cmds, m.handleWorkflowEvent(msg.event))
		cmds = append(cmds, m.operationManager.ListenForEvents())
	}

	if m.width > 0 && m.height > 0 {
		m.applyLayout(m.width, m.height)
	}

	if m.mainScrollEnabled() {
		if needsMainRebuild {
			m.rebuildMainScrollContent()
			if m.snapMainToTop {
				m.mainPane.GotoTop()
				m.snapMainToTop = false
			}
		}
		mainMsg, logMsg := m.splitScrollMsg(msg)
		var mainCmd tea.Cmd
		m.mainPane, mainCmd = m.mainPane.Update(mainMsg)
		if mainCmd != nil {
			cmds = append(cmds, mainCmd)
		}
		_, logCmd := m.logger.Update(logMsg)
		if cmd, ok := logCmd.(tea.Cmd); ok && cmd != nil {
			cmds = append(cmds, cmd)
		}
	} else {
		_, logCmd := m.logger.Update(msg)
		if cmd, ok := logCmd.(tea.Cmd); ok && cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// buildAnswers assembles QuestionAnswer pairs from the current input values.
// Questions with empty answers are included with an empty string.
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

// submitClarifyingAnswers sends answers (possibly nil to skip) back to the
// workflow, transitions to the PRD generation phase, and returns the commands
// needed to keep the event loop running (ListenForEvents must be restarted
// since we returned early from the clarifying key-handler without re-queuing it).
func (m *Model) submitClarifyingAnswers(qas []prompt.QuestionAnswer) []tea.Cmd {
	if m.clarifyAnswersCh != nil {
		m.clarifyAnswersCh <- qas
		m.clarifyAnswersCh = nil
	}
	m.phase = PhasePRDGeneration
	m.logger.AddLog("Clarifications received, generating PRD...")
	// Must restart the event listener — the clarifying key-handler returns early
	// and never reaches the workflowEventMsg branch that normally re-queues it.
	return []tea.Cmd{m.operationManager.ListenForEvents()}
}

func (m *Model) handleWorkflowEvent(event workflow.Event) tea.Cmd {
	switch e := event.(type) {
	case workflow.EventClarifyingQuestions:
		// Convert workflow event to a TUI message so the model can render the form.
		// We must return a command (not modify state directly in handleWorkflowEvent)
		// because this is called from Update's workflowEventMsg branch.
		return func() tea.Msg {
			return clarifyQuestionsMsg{
				questions: e.Questions,
				answersCh: e.AnswersCh,
			}
		}

	case workflow.EventPRDGenerating:
		m.phase = PhasePRDGeneration
		m.logger.AddLog("Generating PRD...")
		m.markMainScrollJump()

	case workflow.EventPRDGenerated:
		m.prd = e.PRD
		m.logger.AddLog(fmt.Sprintf("PRD generated: %s (%d stories)", e.PRD.ProjectName, len(e.PRD.Stories)))
		if m.dryRun {
			m.phase = PhaseCompleted
			m.logger.AddLog("Dry run complete - PRD saved to " + m.cfg.PRDFile)
		} else {
			m.phase = PhasePRDReview
		}
		m.markMainScrollJump()

	case workflow.EventPRDLoaded:
		m.prd = e.PRD
		m.logger.AddLog(fmt.Sprintf("Loaded PRD: %s (%d/%d completed)",
			e.PRD.ProjectName, e.PRD.CompletedCount(), len(e.PRD.Stories)))
		if m.dryRun {
			m.phase = PhaseCompleted
		} else {
			m.phase = PhasePRDReview
		}
		m.markMainScrollJump()

	case workflow.EventPRDReview:
		m.phase = PhasePRDReview
		m.prd = e.PRD
		m.logger.AddLog("PRD ready for review")
		m.markMainScrollJump()

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
		m.err = e.Err
		m.phase = PhaseFailed
		m.markMainScrollJump()

	case workflow.EventCompleted:
		m.phase = PhaseCompleted
		m.logger.AddLog("All stories completed!")
		m.markMainScrollJump()

	case workflow.EventFailed:
		m.phase = PhaseFailed
		if len(e.FailedStories) > 0 {
			m.logger.AddLog(fmt.Sprintf("Failed: %d stories exceeded retry limit", len(e.FailedStories)))
		}
		m.markMainScrollJump()
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
