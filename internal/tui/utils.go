package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) contentWidth(extraIndent int) int {
	width := m.mainPane.Width
	if width <= 0 && m.width > 0 {
		width = max(20, m.width-4)
	}
	if width <= 0 {
		return 76
	}
	return max(20, width-extraIndent)
}

func (m *Model) terminalWidth(extraIndent int) int {
	if m.width > 0 {
		return max(20, m.width-extraIndent)
	}
	return m.contentWidth(extraIndent)
}

func renderStyledWrapped(style lipgloss.Style, text string, width int) string {
	if width <= 0 {
		return style.Render(text)
	}
	return style.Render(wrapText(text, width))
}

func renderIndentedWrapped(style lipgloss.Style, text string, lineWidth int, firstPrefix, continuationPrefix string) string {
	textWidth := max(20, lineWidth-lipgloss.Width(firstPrefix))
	wrapped := wrapText(text, textWidth)
	lines := strings.Split(wrapped, "\n")
	var b strings.Builder
	for i, line := range lines {
		if i > 0 {
			b.WriteString("\n")
		}
		if i == 0 {
			b.WriteString(style.Render(firstPrefix + line))
			continue
		}
		b.WriteString(style.Render(continuationPrefix + line))
	}
	return b.String()
}

func wrapText(s string, width int) string {
	if width <= 0 {
		return s
	}
	wrapped := lipgloss.NewStyle().Width(width).Render(s)
	lines := strings.Split(wrapped, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " ")
	}
	return strings.Join(lines, "\n")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max < 4 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
