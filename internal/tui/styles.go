package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Enhanced color palette with vibrant, modern colors
	primaryColor   = lipgloss.Color("#A855F7") // Vibrant purple
	primaryDark    = lipgloss.Color("#7C3AED") // Darker purple for depth
	successColor   = lipgloss.Color("#10B981") // Emerald green
	errorColor     = lipgloss.Color("#EF4444") // Bright red
	warningColor   = lipgloss.Color("#F59E0B") // Amber
	mutedColor     = lipgloss.Color("#9CA3AF") // Medium gray
	highlightColor = lipgloss.Color("#3B82F6") // Bright blue
	infoColor      = lipgloss.Color("#06B6D4") // Cyan

	// Additional colors for enhanced visual design
	accentColor     = lipgloss.Color("#C084FC") // Light purple accent
	accentLight     = lipgloss.Color("#E9D5FF") // Very light purple
	surfaceColor    = lipgloss.Color("#111827") // Darker surface for better contrast
	surfaceElevated = lipgloss.Color("#1F2937") // Elevated surface
	borderColor     = lipgloss.Color("#4B5563") // Border gray
	borderAccent    = lipgloss.Color("#7C3AED") // Accent border
	textColor       = lipgloss.Color("#F9FAFB") // Bright white text
	textSecondary   = lipgloss.Color("#D1D5DB") // Secondary text
	subtleColor     = lipgloss.Color("#6B7280") // Subtle text

	// Enhanced header with gradient effect
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor).
			Background(primaryColor).
			Padding(0, 3).
			MarginTop(0).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryDark).
			BorderTop(true).
			BorderBottom(true).
			BorderLeft(true).
			BorderRight(true)

	headerTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(textColor).
				Background(primaryColor).
				MarginRight(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(textSecondary).
			Italic(false).
			MarginLeft(1).
			MarginTop(0).
			MarginBottom(0)

	// Enhanced title style with better hierarchy
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor).
			MarginTop(1).
			MarginBottom(1).
			PaddingLeft(1)

	// Enhanced box style with better borders
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderAccent).
			Background(surfaceElevated).
			Foreground(textColor).
			Padding(1, 2).
			MarginBottom(1).
			BorderTop(true).
			BorderBottom(true).
			BorderLeft(true).
			BorderRight(true)

	// Enhanced phase style with more prominence
	phaseStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor).
			Background(surfaceElevated).
			Padding(0, 2).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentColor).
			BorderTop(true).
			BorderBottom(true).
			BorderLeft(true).
			BorderRight(true)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true).
			Background(surfaceElevated)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			Background(surfaceElevated)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true).
			Background(surfaceElevated)

	inProgressStyle = lipgloss.NewStyle().
			Foreground(highlightColor).
			Bold(true).
			Background(surfaceElevated)

	pendingStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Background(surfaceElevated)

	storyItemStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(textColor).
			MarginBottom(1)

	selectedStoryStyle = lipgloss.NewStyle().
				Foreground(textColor).
				Bold(true).
				Background(primaryColor).
				Padding(0, 2).
				MarginBottom(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(accentColor).
				BorderTop(true).
				BorderBottom(true).
				BorderLeft(true).
				BorderRight(true)

	// Enhanced log box with better styling
	logBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Background(surfaceColor).
			Foreground(textColor).
			Padding(1, 2).
			BorderTop(true).
			BorderBottom(true).
			BorderLeft(true).
			BorderRight(true)

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
			Italic(true).
			MarginTop(1).
			MarginBottom(1).
			PaddingLeft(1)

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
