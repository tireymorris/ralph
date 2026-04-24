package tui

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) applyLayout(width, height int) {
	if width == m.layoutSigW && height == m.layoutSigH {
		return
	}
	m.layoutSigW = width
	m.layoutSigH = height
	m.width = width
	m.height = height

	// Reserve ~3 lines for the help bar at the bottom.
	paneHeight := max(4, height-scrollChrome)

	m.mainPane.Width = max(20, width-4)
	m.mainPane.Height = paneHeight
	// Log viewport height must leave room for logBoxStyle border (2) + padding (2).
	m.logger.SetSize(width, paneHeight-4)
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
