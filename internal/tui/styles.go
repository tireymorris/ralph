package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	successColor   = lipgloss.Color("#10B981") // Green
	errorColor     = lipgloss.Color("#EF4444") // Red
	warningColor   = lipgloss.Color("#F59E0B") // Amber
	mutedColor     = lipgloss.Color("#6B7280") // Gray
	highlightColor = lipgloss.Color("#3B82F6") // Blue

	// Base styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// Box styles
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// Status styles
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

	// Story list styles
	storyItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedStoryStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(highlightColor).
				Bold(true)

	// Header styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Padding(0, 2).
			MarginBottom(1)

	// Phase indicator styles
	phaseStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginTop(1).
			MarginBottom(1)

	// Log styles
	logBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(0, 1).
			MarginTop(1)

	logLineStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	logErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor)

	// Help styles
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	// Progress bar styles
	progressFullStyle = lipgloss.NewStyle().
				Foreground(successColor)

	progressEmptyStyle = lipgloss.NewStyle().
				Foreground(mutedColor)
)

// Status icons
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
