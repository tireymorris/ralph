package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
		if m.phase == PhaseClarifying && len(m.clarifyInputs) > 0 {
			switch msg.String() {
			case "ctrl+c":
				m.quitting = true
				m.operationManager.Cancel()
				return m, tea.Quit
			case "esc":
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
					m.clarifyInputs[m.clarifyFocused].Blur()
					m.clarifyFocused++
					m.clarifyInputs[m.clarifyFocused].Focus()
					cmds = append(cmds, textinput.Blink)
				} else {
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

		if m.phase == PhaseFailed && msg.String() == "r" {
			useImpl := m.retryImplementation
			m.retryImplementation = false
			m.err = nil
			m.scrollPane = focusMain
			m.snapMainToTop = true
			var cmd tea.Cmd
			if useImpl && m.prd != nil {
				m.phase = PhaseImplementation
				cmd = tea.Batch(
					m.operationManager.StartImplementation(m.prd),
					m.operationManager.ListenForEvents(),
				)
			} else {
				m.clarifyQuestions = nil
				m.clarifyInputs = nil
				m.clarifyAnswersCh = nil
				m.clarifyFocused = 0
				m.phase = PhasePRDGeneration
				cmd = tea.Batch(
					m.operationManager.StartFullOperation(m.resume, m.prompt),
					m.operationManager.ListenForEvents(),
				)
			}
			if m.width > 0 && m.height > 0 {
				m.applyLayout(m.width, m.height)
			}
			m.rebuildMainScrollContent()
			return m, cmd
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
