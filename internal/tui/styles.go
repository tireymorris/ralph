package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

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
			Foreground(textSecondary).
			Padding(1, 2)

	clarifyBodyStyle = lipgloss.NewStyle().
				PaddingLeft(2)

	clarifyQuestionStyle = lipgloss.NewStyle().
				Foreground(subtleColor)

	clarifyInputFocusedStyle = lipgloss.NewStyle().
					Foreground(subtleColor).
					PaddingLeft(2).
					Border(lipgloss.ThickBorder()).
					BorderForeground(primaryColor).
					BorderTop(false).
					BorderBottom(false).
					BorderLeft(true).
					BorderRight(false)

	clarifyInputStyle = lipgloss.NewStyle().
				Foreground(subtleColor).
				PaddingLeft(4)

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
	iconSuccess    = "✓"
	iconWarning    = "⚠"
)

func configureTextInput(ti textinput.Model) textinput.Model {
	ti.PromptStyle = lipgloss.NewStyle().Foreground(subtleColor)
	ti.TextStyle = lipgloss.NewStyle().Foreground(subtleColor)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(mutedColor)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(primaryColor)
	ti.Cursor.TextStyle = lipgloss.NewStyle().Foreground(subtleColor)
	return ti
}

func getStatusIcon(passes bool, inProgress bool) string {
	if passes {
		return successStyle.Render(iconCompleted)
	}
	if inProgress {
		return inProgressStyle.Render(iconInProgress)
	}
	return pendingStyle.Render(iconPending)
}

func getStatusText(passes bool, inProgress bool) string {
	if passes {
		return successStyle.Render("completed")
	}
	if inProgress {
		return inProgressStyle.Render("in progress")
	}
	return pendingStyle.Render("pending")
}
