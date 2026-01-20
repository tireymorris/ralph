package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n")
	b.WriteString(m.renderPhase())
	b.WriteString("\n")

	switch m.phase {
	case PhaseInit, PhasePRDGeneration:
		b.WriteString(m.renderGenerating())
	case PhaseImplementation:
		b.WriteString(m.renderImplementation())
	case PhaseCompleted:
		b.WriteString(m.renderCompleted())
	case PhaseFailed:
		b.WriteString(m.renderFailed())
	}

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("Output Logs"))
	b.WriteString("\n")
	b.WriteString(m.renderLogs())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ scroll logs • q quit • ctrl+c exit"))

	return b.String()
}

func (m *Model) renderHeader() string {
	title := headerTitleStyle.Render("RALPH")
	subtitle := subtitleStyle.Render("Autonomous software development agent")
	return headerStyle.Render(title + subtitle)
}

func (m *Model) renderPhase() string {
	icon := m.spinner.View()
	if m.phase == PhaseCompleted {
		icon = iconSuccess
	} else if m.phase == PhaseFailed {
		icon = iconFailed
	}
	return phaseStyle.Render(fmt.Sprintf("%s %s", icon, m.phase.String()))
}

func (m *Model) renderGenerating() string {
	promptLabel := labelStyle.Render("Prompt")
	promptTextStyle := lipgloss.NewStyle().Foreground(textColor)
	promptText := promptTextStyle.Render(truncate(m.prompt, 60))
	generatingText := inProgressStyle.Render("Generating PRD from your requirements...")

	content := fmt.Sprintf("%s %s\n\n%s %s", promptLabel, promptText, m.spinner.View(), generatingText)
	return infoStyle.Render(content)
}

func (m *Model) renderImplementation() string {
	if m.prd == nil {
		return ""
	}

	var b strings.Builder

	// Project info section - clean lines without box
	projectLabel := labelStyle.Render("Project")
	projectValue := valueStyle.Render(m.prd.ProjectName)
	b.WriteString(infoStyle.Render(projectLabel + " " + projectValue))
	b.WriteString("\n")

	if m.prd.BranchName != "" {
		branchLabel := labelStyle.Render("Branch")
		branchValue := valueStyle.Render(m.prd.BranchName)
		b.WriteString(infoStyle.Render(branchLabel + " " + branchValue))
		b.WriteString("\n")
	}

	// Progress section - standalone progress bar
	completed := m.prd.CompletedCount()
	total := len(m.prd.Stories)
	percent := float64(completed) / float64(total)

	progressLabel := labelStyle.Render("Progress")
	progressValue := mutedStyle.Render(fmt.Sprintf("%d/%d stories", completed, total))
	b.WriteString(infoStyle.Render(progressLabel + " " + progressValue))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render(m.progress.ViewAs(percent)))
	b.WriteString("\n\n")

	// Stories section
	b.WriteString(titleStyle.Render("Stories"))
	b.WriteString("\n")
	for _, s := range m.prd.Stories {
		isCurrentStory := m.currentStory != nil && s.ID == m.currentStory.ID
		icon := getStatusIcon(s.Passes, isCurrentStory, s.RetryCount, m.cfg.RetryAttempts)
		status := getStatusText(s.Passes, isCurrentStory, s.RetryCount, m.cfg.RetryAttempts)

		if isCurrentStory {
			line := fmt.Sprintf("%s %s  %s", icon, s.Title, status)
			b.WriteString(selectedStoryStyle.Render(line))
		} else {
			line := fmt.Sprintf("%s %s  %s", icon, s.Title, status)
			b.WriteString(storyItemStyle.Render(line))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m *Model) renderCompleted() string {
	var b strings.Builder

	if m.dryRun {
		b.WriteString(successStyle.Render(iconSuccess + " Dry run completed!"))
		b.WriteString("\n\n")
		b.WriteString(labelStyle.Render("PRD saved to") + " " + valueStyle.Render(m.cfg.PRDFile))
		b.WriteString("\n")
		b.WriteString(mutedStyle.Render("Run without --dry-run to implement, or use --resume."))
		b.WriteString("\n")
	} else if m.prd != nil {
		b.WriteString(successStyle.Render(iconSuccess + " All stories completed!"))
		b.WriteString("\n\n")
		b.WriteString(labelStyle.Render("Project") + " " + valueStyle.Render(m.prd.ProjectName))
		b.WriteString("\n")
		b.WriteString(labelStyle.Render("Stories") + " " + valueStyle.Render(fmt.Sprintf("%d completed", len(m.prd.Stories))))
		b.WriteString("\n")
		b.WriteString(labelStyle.Render("Iterations") + " " + valueStyle.Render(fmt.Sprintf("%d", m.iteration)))
		b.WriteString("\n")
	}

	return infoStyle.Render(b.String())
}

func (m *Model) renderFailed() string {
	var b strings.Builder

	b.WriteString(errorStyle.Render(iconFailed + " Implementation failed"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(labelStyle.Render("Error") + " " + errorStyle.Render(fmt.Sprintf("%v", m.err)))
		b.WriteString("\n")
	}

	if m.prd != nil {
		failed := m.prd.FailedStories(m.cfg.RetryAttempts)
		if len(failed) > 0 {
			b.WriteString("\n")
			b.WriteString(warningStyle.Render(fmt.Sprintf("%s Failed stories (%d):", iconWarning, len(failed))))
			b.WriteString("\n")
			for _, s := range failed {
				b.WriteString(fmt.Sprintf("    %s %s (%d attempts)\n", iconFailed, s.Title, s.RetryCount))
			}
		}
		b.WriteString("\n")
		b.WriteString(mutedStyle.Render("Run with --resume to retry after fixing issues."))
		b.WriteString("\n")
	}

	return infoStyle.Render(b.String())
}

func (m *Model) renderLogs() string {
	viewportContent := m.logger.GetView().View()
	if viewportContent == "" {
		return logBoxStyle.Render(mutedStyle.Render("Waiting for output..."))
	}
	return logBoxStyle.Render(viewportContent)
}
