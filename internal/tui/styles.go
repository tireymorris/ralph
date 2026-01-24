package tui

import "github.com/charmbracelet/lipgloss"

var (
	primaryColor   = lipgloss.Color("#A855F7")
	successColor   = lipgloss.Color("#10B981")
	errorColor     = lipgloss.Color("#EF4444")
	warningColor   = lipgloss.Color("#F59E0B")
	mutedColor     = lipgloss.Color("#9CA3AF")
	highlightColor = lipgloss.Color("#3B82F6")
	infoColor      = lipgloss.Color("#06B6D4")
	accentColor    = lipgloss.Color("#C084FC")
	surfaceColor   = lipgloss.Color("#111827")
	borderColor    = lipgloss.Color("#4B5563")
	textColor      = lipgloss.Color("#F9FAFB")
	textSecondary  = lipgloss.Color("#D1D5DB")
	subtleColor    = lipgloss.Color("#6B7280")

	headerStyle = lipgloss.NewStyle().
			MarginTop(1).
			MarginBottom(1)

	headerTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(textColor).
				Background(primaryColor).
				Padding(0, 2)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginLeft(1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor).
			MarginTop(1).
			MarginBottom(1).
			PaddingLeft(2)

	infoStyle = lipgloss.NewStyle().
			Foreground(textColor).
			PaddingLeft(2)

	labelStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	valueStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Bold(true)

	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Foreground(textColor).
			Padding(1, 2).
			MarginBottom(1)

	phaseStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor).
			PaddingLeft(2).
			MarginBottom(1).
			Border(lipgloss.ThickBorder()).
			BorderForeground(primaryColor).
			BorderTop(false).
			BorderBottom(false).
			BorderLeft(true).
			BorderRight(false)

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
			Foreground(subtleColor)

	storyItemStyle = lipgloss.NewStyle().
			PaddingLeft(4).
			Foreground(textSecondary)

	selectedStoryStyle = lipgloss.NewStyle().
				Foreground(textColor).
				Bold(true).
				PaddingLeft(2).
				Border(lipgloss.ThickBorder()).
				BorderForeground(primaryColor).
				BorderTop(false).
				BorderBottom(false).
				BorderLeft(true).
				BorderRight(false)

	logBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Background(surfaceColor).
			Foreground(textSecondary).
			Padding(1, 2)

	logLineStyle = lipgloss.NewStyle().
			Foreground(textSecondary).
			PaddingLeft(1)

	logErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			PaddingLeft(1)

	logSuccessStyle = lipgloss.NewStyle().
			Foreground(successColor).
			PaddingLeft(1)

	logInfoStyle = lipgloss.NewStyle().
			Foreground(infoColor).
			PaddingLeft(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(subtleColor).
			MarginTop(1).
			PaddingLeft(2)

)

const (
	iconPending    = "○"
	iconInProgress = "◐"
	iconCompleted  = "●"
	iconFailed     = "✗"
	iconSuccess    = "✓"
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
