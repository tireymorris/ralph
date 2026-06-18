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

func continuationAfterIcon(icon string) string {
	return strings.Repeat(" ", lipgloss.Width(icon)+1)
}

func fitStatusOnLastLine(lines []string, firstPrefix, continuationPrefix, statusSuffix string, lineWidth int) []string {
	if len(lines) == 0 {
		return lines
	}

	lastIdx := len(lines) - 1
	prefix := continuationPrefix
	if lastIdx == 0 {
		prefix = firstPrefix
	}

	lastLine := lines[lastIdx]
	if lipgloss.Width(prefix+lastLine+statusSuffix) <= lineWidth {
		lines[lastIdx] = lastLine + statusSuffix
		return lines
	}

	available := max(20, lineWidth-lipgloss.Width(prefix)-lipgloss.Width(statusSuffix))
	if lipgloss.Width(lastLine) <= available {
		lines[lastIdx] = lastLine + statusSuffix
		return lines
	}

	sublines := strings.Split(wrapText(lastLine, available), "\n")
	if len(sublines) == 1 {
		lines[lastIdx] = lastLine + statusSuffix
		return lines
	}

	result := make([]string, 0, lastIdx+len(sublines))
	result = append(result, lines[:lastIdx]...)
	result = append(result, sublines[:len(sublines)-1]...)
	result = append(result, sublines[len(sublines)-1]+statusSuffix)
	return result
}

func renderStatusWrapped(style lipgloss.Style, firstPrefix, text, status string, lineWidth int, continuationPrefix string) string {
	statusSuffix := "  " + status
	textWidth := max(20, lineWidth-lipgloss.Width(continuationPrefix))
	lines := strings.Split(wrapText(text, textWidth), "\n")
	lines = fitStatusOnLastLine(lines, firstPrefix, continuationPrefix, statusSuffix, lineWidth)

	var b strings.Builder
	for i, line := range lines {
		if i > 0 {
			b.WriteString("\n")
		}
		prefix := firstPrefix
		if i > 0 {
			prefix = continuationPrefix
		}
		b.WriteString(style.Render(prefix + line))
	}
	return b.String()
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
