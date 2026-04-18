package tui

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) applyLayout(width, height int) {
	lc := m.logger.LogCount()
	bias := m.logHeightBias

	if m.fullscreenPane != focusNone {
		if m.fullscreenPane == focusMain {
			m.logger.SetSize(width, 0)
			m.mainPane.Width = max(20, width-4)
			m.mainPane.Height = max(4, height-scrollChrome)
		} else {
			m.logger.SetSize(width, max(4, height-scrollChrome))
			m.mainPane.Width = max(20, width-4)
			m.mainPane.Height = 0
		}
		m.layoutSigW = width
		m.layoutSigH = height
		m.layoutSigLogCount = lc
		m.layoutSigBias = bias
		m.width = width
		m.height = height
		return
	}

	mainH, logH := computePaneHeights(height, lc, bias, m.fullscreenPane)
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
