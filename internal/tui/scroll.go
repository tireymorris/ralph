package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type scrollFocus int

const (
	focusMain scrollFocus = iota
	focusLogs
)

var noopScrollMsg = struct{}{}

var scrollKeyMap = viewport.DefaultKeyMap()

func isViewportScrollKey(msg tea.KeyMsg) bool {
	km := scrollKeyMap
	return key.Matches(msg, km.PageDown) ||
		key.Matches(msg, km.PageUp) ||
		key.Matches(msg, km.HalfPageUp) ||
		key.Matches(msg, km.HalfPageDown) ||
		key.Matches(msg, km.Down) ||
		key.Matches(msg, km.Up) ||
		key.Matches(msg, km.Left) ||
		key.Matches(msg, km.Right)
}

func isScrollNavMsg(msg tea.Msg) bool {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return isViewportScrollKey(msg)
	case tea.MouseMsg:
		if msg.Action != tea.MouseActionPress {
			return false
		}
		return msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown
	default:
		return false
	}
}
