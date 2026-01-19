package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Modern color palette with better contrast and visual appeal
	primaryColor   = lipgloss.Color("#8B5CF6") // Softer purple
	successColor   = lipgloss.Color("#34D399") // Brighter green
	errorColor     = lipgloss.Color("#F87171") // Softer red
	warningColor   = lipgloss.Color("#FBBF24") // Warmer yellow
	mutedColor     = lipgloss.Color("#9CA3AF") // Lighter gray
	highlightColor = lipgloss.Color("#60A5FA") // Lighter blue

	// Additional colors for enhanced visual design
	accentColor  = lipgloss.Color("#A78BFA") // Light purple accent
	surfaceColor = lipgloss.Color("#1F2937") // Dark surface
	borderColor  = lipgloss.Color("#374151") // Border gray
	textColor    = lipgloss.Color("#F3F4F6") // Light text
	subtleColor  = lipgloss.Color("#6B7280") // Subtle text

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginTop(1).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(subtleColor).
			Italic(true).
			MarginLeft(2).
			MarginBottom(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Background(surfaceColor).
			Foreground(textColor).
			Padding(1, 2).
			MarginBottom(1)

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
			PaddingLeft(2).
			Foreground(textColor).
			MarginBottom(1)

	selectedStoryStyle = lipgloss.NewStyle().
				Foreground(highlightColor).
				Bold(true).
				Background(accentColor).
				Padding(0, 2).
				MarginBottom(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(highlightColor)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor).
			Background(primaryColor).
			Padding(0, 2)

	phaseStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor).
			Background(surfaceColor).
			Padding(0, 1).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentColor)

	logBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Background(surfaceColor).
			Foreground(textColor).
			Padding(1, 2)

	logLineStyle = lipgloss.NewStyle().
			Foreground(subtleColor).
			PaddingLeft(1)

	logErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			PaddingLeft(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(subtleColor).
			Italic(true).
			MarginTop(1).
			MarginBottom(1)

	progressFullStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Bold(true)

	progressEmptyStyle = lipgloss.NewStyle().
				Foreground(borderColor)
)

const (
	iconPending    = "○"
	iconInProgress = "◐"
	iconCompleted  = "●"
	iconFailed     = "✗"
	iconSuccess    = "✓"
	iconWorking    = "⚡"
	iconWarning    = "⚠"
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
