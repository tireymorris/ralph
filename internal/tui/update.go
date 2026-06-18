package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.spinner.Tick,
		m.operationManager.ListenForEvents(),
		tea.WindowSize(),
	}
	if m.prompt == "" && !m.resume {
		m.phase = PhaseAwaitingPrompt
		m.promptInput.Focus()
		cmds = append(cmds, textinput.Blink)
	} else {
		cmds = append(cmds, m.operationManager.StartFullOperation(m.resume, m.prompt))
	}
	return tea.Batch(cmds...)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var needsMainRebuild bool

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.phase == PhaseAwaitingPrompt {
			switch msg.String() {
			case "q", "ctrl+c":
				m.quitting = true
				m.operationManager.Cancel()
				return m, tea.Quit
			case "enter":
				trimmed := strings.TrimSpace(m.promptInput.Value())
				if len(trimmed) >= 1 {
					m.prompt = trimmed
					m.phase = PhasePRDGeneration
					return m, m.operationManager.StartFullOperation(false, m.prompt)
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.promptInput, cmd = m.promptInput.Update(msg)
				return m, cmd
			}
		}

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

		if m.phase == PhasePRDReview {
			if m.critiqueActive {
				switch msg.String() {
				case "esc":
					m.critiqueActive = false
					m.critiqueInput.SetValue("")
					m.scrollPane = focusMain
					m.rebuildMainScrollContent()
					return m, nil
				case "enter":
					critique := strings.TrimSpace(m.critiqueInput.Value())
					m.critiqueActive = false
					m.critiqueInput.SetValue("")
					m.scrollPane = focusMain
					if critique != "" {
						m.revisingPRD = true
						m.phase = PhasePRDGeneration
						if m.width > 0 && m.height > 0 {
							m.applyLayout(m.width, m.height)
						}
						m.rebuildMainScrollContent()
						m.mainPane.GotoTop()
						return m, tea.Batch(
							m.operationManager.StartCritiqueRevision(m.prompt, critique),
							m.operationManager.ListenForEvents(),
						)
					}
					m.phase = PhaseImplementation
					if m.width > 0 && m.height > 0 {
						m.applyLayout(m.width, m.height)
					}
					m.rebuildMainScrollContent()
					m.mainPane.GotoTop()
					return m, tea.Batch(
						m.operationManager.ApproveReview(),
						m.operationManager.ListenForEvents(),
					)
				default:
					var cmd tea.Cmd
					m.critiqueInput, cmd = m.critiqueInput.Update(msg)
					cmds = append(cmds, cmd)
				}
				m.rebuildMainScrollContent()
				return m, tea.Batch(cmds...)
			}
			if msg.String() == "c" {
				m.critiqueActive = true
				m.scrollPane = focusMain
				m.critiqueInput.Focus()
				cmds = append(cmds, textinput.Blink)
				m.rebuildMainScrollContent()
				return m, tea.Batch(cmds...)
			}
			if msg.String() == "enter" {
				m.phase = PhaseImplementation
				m.scrollPane = focusMain
				if m.width > 0 && m.height > 0 {
					m.applyLayout(m.width, m.height)
				}
				m.rebuildMainScrollContent()
				m.mainPane.GotoTop()
				return m, tea.Batch(
					m.operationManager.ApproveReview(),
					m.operationManager.ListenForEvents(),
				)
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

	case resumeStartMsg:
		m.phase = msg.phase
		m.prd = msg.prd
		m.snapshot = msg.snapshot
		needsMainRebuild = true

	case operationErrorMsg:
		m.err = msg.err
		m.phase = PhaseFailed
		needsMainRebuild = true

	case clarifyQuestionsMsg:
		m.phase = PhaseClarifying
		m.clarifyQuestions = msg.questions
		m.clarifyAnswersCh = msg.answersCh
		m.clarifyFocused = 0
		m.clarifyInputs = make([]textinput.Model, len(msg.questions))
		for i := range m.clarifyInputs {
			ti := configureTextInput(textinput.New())
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
