package tui

import "github.com/charmbracelet/lipgloss"

var (
	primaryColor   = lipgloss.Color("#7C3AED")
	successColor   = lipgloss.Color("#10B981")
	errorColor     = lipgloss.Color("#EF4444")
	warningColor   = lipgloss.Color("#F59E0B")
	mutedColor     = lipgloss.Color("#6B7280")
	highlightColor = lipgloss.Color("#3B82F6")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	inProgressStyle = lipgloss.NewStyle().
			Foreground(highlightColor).
			Bold(true)

	pendingStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	storyItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedStoryStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(highlightColor).
				Bold(true)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Padding(0, 2).
			MarginBottom(1)

	phaseStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginTop(1).
			MarginBottom(1)

	logBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(0, 1).
			MarginTop(1)

	logLineStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	logErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor)

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	progressFullStyle = lipgloss.NewStyle().
				Foreground(successColor)

	progressEmptyStyle = lipgloss.NewStyle().
				Foreground(mutedColor)
)

const (
	iconPending    = "○"
	iconInProgress = "◐"
	iconCompleted  = "●"
	iconFailed     = "✗"
)

func getStatusIcon(passes bool, inProgress bool, retryCount, maxRetries int) string {
	if passes {
		return successStyle.Render(iconCompleted)
	}
	if inProgress {
		return inProgressStyle.Render(iconInProgress)
	}
	if retryCount >= maxRetries {
		return errorStyle.Render(iconFailed)
	}
	return pendingStyle.Render(iconPending)
}

func getStatusText(passes bool, inProgress bool, retryCount, maxRetries int) string {
	if passes {
		return successStyle.Render("completed")
	}
	if inProgress {
		return inProgressStyle.Render("in progress")
	}
	if retryCount >= maxRetries {
		return errorStyle.Render("failed")
	}
	return pendingStyle.Render("pending")
}
