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
	b.WriteString("\n\n")

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

	b.WriteString("\n\n")
	b.WriteString(titleStyle.Render("📝 Output Logs"))
	b.WriteString("\n")
	b.WriteString(m.renderLogs())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ scroll logs • q quit • ctrl+c exit"))

	return b.String()
}

func (m *Model) renderHeader() string {
	return headerStyle.Render("⚡ RALPH") + subtitleStyle.Render("Autonomous software development agent")
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
	return boxStyle.Render(fmt.Sprintf(
		"📝 Prompt: %s\n\n⚡ Generating PRD from your requirements...",
		truncate(m.prompt, 60),
	))
}

func (m *Model) renderImplementation() string {
	if m.prd == nil {
		return ""
	}

	var b strings.Builder

	// Project info section
	projectInfo := fmt.Sprintf("📁 Project: %s", m.prd.ProjectName)
	if m.prd.BranchName != "" {
		projectInfo += fmt.Sprintf(" | 🌿 Branch: %s", m.prd.BranchName)
	}
	b.WriteString(boxStyle.Render(projectInfo))
	b.WriteString("\n")

	// Progress section
	completed := m.prd.CompletedCount()
	total := len(m.prd.Stories)
	percent := float64(completed) / float64(total)

	progressSection := fmt.Sprintf("📊 Progress: %d/%d stories completed", completed, total)
	b.WriteString(boxStyle.Render(progressSection + "\n" + m.progress.ViewAs(percent)))
	b.WriteString("\n")

	// Stories section
	b.WriteString(titleStyle.Render("📋 Stories"))
	for _, s := range m.prd.Stories {
		isCurrentStory := m.currentStory != nil && s.ID == m.currentStory.ID
		icon := getStatusIcon(s.Passes, isCurrentStory, s.RetryCount, m.cfg.RetryAttempts)
		status := getStatusText(s.Passes, isCurrentStory, s.RetryCount, m.cfg.RetryAttempts)

		var prefix string
		if isCurrentStory {
			prefix = iconWorking + " "
		}

		line := fmt.Sprintf("%s%s %s [%s]", prefix, icon, s.Title, status)
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
		b.WriteString(successStyle.Render(iconSuccess + " Dry run completed!\n\n"))
		b.WriteString(fmt.Sprintf("📄 PRD saved to: %s\n", m.cfg.PRDFile))
		b.WriteString("💡 Run without --dry-run to implement, or use --resume.\n")
	} else if m.prd != nil {
		b.WriteString(successStyle.Render(iconSuccess + " All stories completed!\n\n"))
		b.WriteString(fmt.Sprintf("📁 Project: %s\n", m.prd.ProjectName))
		b.WriteString(fmt.Sprintf("📊 Stories: %d completed\n", len(m.prd.Stories)))
		b.WriteString(fmt.Sprintf("🔄 Iterations: %d\n", m.iteration))
	}

	return boxStyle.Render(b.String())
}

func (m *Model) renderFailed() string {
	var b strings.Builder

	b.WriteString(errorStyle.Render(iconFailed + " Implementation failed\n\n"))

	if m.err != nil {
		b.WriteString(fmt.Sprintf("❌ Error: %v\n", m.err))
	}

	if m.prd != nil {
		failed := m.prd.FailedStories(m.cfg.RetryAttempts)
		if len(failed) > 0 {
			b.WriteString(fmt.Sprintf("\n%s Failed stories (%d):\n", iconWarning, len(failed)))
			for _, s := range failed {
				b.WriteString(fmt.Sprintf("  • %s (%d attempts)\n", s.Title, s.RetryCount))
			}
		}
		b.WriteString("\n💡 Run with --resume to retry after fixing issues.\n")
	}

	return boxStyle.Render(b.String())
}

func (m *Model) renderLogs() string {
	viewportContent := m.logger.GetView().View()
	if viewportContent == "" {
		return logBoxStyle.Render("📋 Waiting for output...")
	}
	return logBoxStyle.Render(viewportContent)
}
