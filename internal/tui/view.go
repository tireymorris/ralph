package tui

import (
	"fmt"
	"strings"
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
	b.WriteString(m.renderLogs())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press q to quit"))

	return b.String()
}

func (m *Model) renderHeader() string {
	return headerStyle.Render("🤖 RALPH - Autonomous Software Development Agent")
}

func (m *Model) renderPhase() string {
	icon := m.spinner.View()
	if m.phase == PhaseCompleted {
		icon = "✓"
	} else if m.phase == PhaseFailed {
		icon = "✗"
	}
	return phaseStyle.Render(fmt.Sprintf("%s %s", icon, m.phase.String()))
}

func (m *Model) renderGenerating() string {
	return boxStyle.Render(fmt.Sprintf(
		"Prompt: %s\n\nGenerating PRD from your requirements...",
		truncate(m.prompt, 60),
	))
}

func (m *Model) renderImplementation() string {
	if m.prd == nil {
		return ""
	}

	var b strings.Builder

	b.WriteString(fmt.Sprintf("📁 Project: %s\n", m.prd.ProjectName))
	if m.prd.BranchName != "" {
		b.WriteString(fmt.Sprintf("🌿 Branch: %s\n", m.prd.BranchName))
	}
	b.WriteString("\n")

	completed := m.prd.CompletedCount()
	total := len(m.prd.Stories)
	percent := float64(completed) / float64(total)

	b.WriteString(fmt.Sprintf("Progress: %d/%d stories ", completed, total))
	b.WriteString(m.progress.ViewAs(percent))
	b.WriteString("\n\n")

	b.WriteString("Stories:\n")
	for _, s := range m.prd.Stories {
		isCurrentStory := m.currentStory != nil && s.ID == m.currentStory.ID
		icon := getStatusIcon(s.Passes, isCurrentStory, s.RetryCount, m.cfg.RetryAttempts)
		status := getStatusText(s.Passes, isCurrentStory, s.RetryCount, m.cfg.RetryAttempts)

		line := fmt.Sprintf("%s %s [%s]", icon, s.Title, status)
		if isCurrentStory {
			b.WriteString(selectedStoryStyle.Render(line))
		} else {
			b.WriteString(storyItemStyle.Render(line))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m *Model) renderCompleted() string {
	var b strings.Builder

	if m.dryRun {
		b.WriteString(successStyle.Render("✓ Dry run completed!\n\n"))
		b.WriteString(fmt.Sprintf("PRD saved to: %s\n", m.cfg.PRDFile))
		b.WriteString("Run without --dry-run to implement, or use --resume.\n")
	} else if m.prd != nil {
		b.WriteString(successStyle.Render("✓ All stories completed!\n\n"))
		b.WriteString(fmt.Sprintf("📁 Project: %s\n", m.prd.ProjectName))
		b.WriteString(fmt.Sprintf("📊 Stories: %d completed\n", len(m.prd.Stories)))
		b.WriteString(fmt.Sprintf("📝 Iterations: %d\n", m.iteration))
	}

	return boxStyle.Render(b.String())
}

func (m *Model) renderFailed() string {
	var b strings.Builder

	b.WriteString(errorStyle.Render("✗ Implementation failed\n\n"))

	if m.err != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n", m.err))
	}

	if m.prd != nil {
		failed := m.prd.FailedStories(m.cfg.RetryAttempts)
		if len(failed) > 0 {
			b.WriteString(fmt.Sprintf("\nFailed stories (%d):\n", len(failed)))
			for _, s := range failed {
				b.WriteString(fmt.Sprintf("  • %s (%d attempts)\n", s.Title, s.RetryCount))
			}
		}
		b.WriteString("\nRun with --resume to retry after fixing issues.\n")
	}

	return boxStyle.Render(b.String())
}

func (m *Model) renderLogs() string {
	if len(m.logs) == 0 {
		return logBoxStyle.Render("Waiting for output...")
	}

	// Show more lines based on terminal height, minimum 12 lines
	maxLines := 12
	if m.height > 40 {
		maxLines = min(20, m.height/3)
	}

	startIdx := 0
	if len(m.logs) > maxLines {
		startIdx = len(m.logs) - maxLines
	}

	var lines []string
	for i := startIdx; i < len(m.logs); i++ {
		lines = append(lines, logLineStyle.Render(truncate(m.logs[i], m.width-6)))
	}

	return logBoxStyle.Render(strings.Join(lines, "\n"))
}
