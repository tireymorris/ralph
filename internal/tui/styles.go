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

	// Compact header badge style (no border)
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

	// Section title style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor).
			MarginTop(1).
			MarginBottom(1).
			PaddingLeft(2)

	// Clean info style without border (for project info, progress)
	infoStyle = lipgloss.NewStyle().
			Foreground(textColor).
			PaddingLeft(2)

	// Label and value styles for key-value pairs
	labelStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	valueStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Bold(true)

	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Box style kept only for content that needs visual containment
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Foreground(textColor).
			Padding(1, 2).
			MarginBottom(1)

	// Clean phase style with left accent border
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

	// Log box - keep bordered for scrollable content
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
