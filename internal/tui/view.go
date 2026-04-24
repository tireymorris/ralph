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

	if m.phase == PhaseClarifying {
		b.WriteString(m.renderHeader())
		b.WriteString("\n")
		b.WriteString(m.renderPhase())
		b.WriteString("\n")
		b.WriteString(m.renderClarifying())
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("Tab/↑/↓ navigate • Enter confirm • Esc skip all • ctrl+c exit"))
		return b.String()
	}

	if m.scrollPane == focusMain {
		b.WriteString(m.mainPane.View())
	} else {
		b.WriteString(titleStyle.Render("Output Logs"))
		b.WriteString("\n")
		b.WriteString(m.renderLogs())
	}

	b.WriteString("\n")
	if m.phase == PhasePRDReview {
		b.WriteString(helpStyle.Render("Tab switch pane • ↑/↓ scroll • Enter continue • q quit • ctrl+c exit"))
	} else if m.phase == PhaseFailed {
		b.WriteString(helpStyle.Render("Tab switch pane • ↑/↓ scroll • r retry • q quit • ctrl+c exit"))
	} else {
		b.WriteString(helpStyle.Render("Tab switch pane • ↑/↓ scroll • q quit • ctrl+c exit"))
	}

	return b.String()
}

func (m *Model) renderHeader() string {
	title := headerTitleStyle.Render("RALPH")
	subtitle := subtitleStyle.Render("Autonomous software development agent")
	return headerStyle.Render(title + subtitle)
}

func (m *Model) renderPhase() string {
	icon := m.spinner.View()
	switch m.phase {
	case PhaseCompleted:
		icon = iconSuccess
	case PhaseClarifying:
		icon = "?"
	case PhasePRDReview:
		icon = "!"
	case PhaseFailed:
		icon = iconWarning
	}
	return phaseStyle.Render(fmt.Sprintf("%s %s", icon, m.phase.String()))
}

func (m *Model) renderClarifying() string {
	if len(m.clarifyQuestions) == 0 {
		return infoStyle.Render(mutedStyle.Render("Waiting for questions..."))
	}

	var b strings.Builder

	b.WriteString(infoStyle.Render(inProgressStyle.Render("Please answer the following questions before we generate your PRD.")))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("  Tab/↑/↓ to navigate  •  Enter to confirm  •  Esc to skip all questions"))
	b.WriteString("\n\n")

	for i, q := range m.clarifyQuestions {
		num := labelStyle.Render(fmt.Sprintf("Q%d.", i+1))
		question := valueStyle.Render(q)
		b.WriteString(infoStyle.Render(fmt.Sprintf("%s %s", num, question)))
		b.WriteString("\n")

		if i < len(m.clarifyInputs) {
			inputView := m.clarifyInputs[i].View()
			if i == m.clarifyFocused {
				b.WriteString(infoStyle.Render(selectedStoryStyle.Render(inputView)))
			} else {
				b.WriteString(infoStyle.Render(storyItemStyle.Render(inputView)))
			}
		}
		b.WriteString("\n\n")
	}

	lastQ := len(m.clarifyQuestions) - 1
	if m.clarifyFocused >= lastQ {
		b.WriteString(infoStyle.Render(successStyle.Render("Press Enter to submit and generate PRD")))
	} else {
		remaining := lastQ - m.clarifyFocused
		b.WriteString(infoStyle.Render(mutedStyle.Render(fmt.Sprintf("%d question(s) remaining", remaining))))
	}

	return b.String()
}

func (m *Model) renderFailed() string {
	msg := "Workflow stopped."
	if m.err != nil {
		msg = m.err.Error()
	}
	return errorStyle.Render(msg)
}

func (m *Model) renderGenerating() string {
	promptLabel := labelStyle.Render("Prompt")
	promptTextStyle := lipgloss.NewStyle().Foreground(textColor)
	promptText := promptTextStyle.Render(truncate(m.prompt, 60))
	generatingText := inProgressStyle.Render("Generating PRD from your requirements...")

	content := fmt.Sprintf("%s %s\n\n%s %s", promptLabel, promptText, m.spinner.View(), generatingText)
	return infoStyle.Render(content)
}

func (m *Model) renderPRDReview() string {
	if m.prd == nil {
		return ""
	}

	var b strings.Builder

	b.WriteString(infoStyle.Render(inProgressStyle.Render("PRD ready for review")))
	b.WriteString("\n\n")

	projectLabel := labelStyle.Render("Project")
	projectValue := valueStyle.Render(m.prd.ProjectName)
	b.WriteString(infoStyle.Render(projectLabel + " " + projectValue))
	b.WriteString("\n")

	branchLabel := labelStyle.Render("Branch")
	branchValue := valueStyle.Render(m.prd.BranchName)
	b.WriteString(infoStyle.Render(branchLabel + " " + branchValue))
	b.WriteString("\n\n")

	b.WriteString(titleStyle.Render("Stories"))
	b.WriteString("\n")
	for _, s := range m.prd.Stories {
		status := "[ ]"
		if s.Passes {
			status = "[x]"
		}
		deps := ""
		if len(s.DependsOn) > 0 {
			deps = " (depends: " + strings.Join(s.DependsOn, ", ") + ")"
		}
		line := fmt.Sprintf("%s P%d %s%s", status, s.Priority, s.Title, deps)
		b.WriteString(storyItemStyle.Render(line))
		b.WriteString("\n")

		if len(s.AcceptanceCriteria) > 0 {
			b.WriteString(mutedStyle.Render("    Acceptance criteria:"))
			b.WriteString("\n")
			for _, ac := range s.AcceptanceCriteria {
				b.WriteString(mutedStyle.Render(fmt.Sprintf("      - %s", ac)))
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Press Enter to continue implementation"))

	return b.String()
}

func (m *Model) renderImplementation() string {
	if m.prd == nil {
		return ""
	}

	var b strings.Builder

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

	completed := m.prd.CompletedCount()
	total := len(m.prd.Stories)
	percent := float64(completed) / float64(total)

	progressLabel := labelStyle.Render("Progress")
	progressValue := mutedStyle.Render(fmt.Sprintf("%d/%d stories", completed, total))
	b.WriteString(infoStyle.Render(progressLabel + " " + progressValue))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render(m.progress.ViewAs(percent)))
	b.WriteString("\n\n")

	b.WriteString(titleStyle.Render("Stories"))
	b.WriteString("\n")
	for _, s := range m.prd.Stories {
		isCurrentStory := m.currentStory != nil && s.ID == m.currentStory.ID
		icon := getStatusIcon(s.Passes, isCurrentStory)
		status := getStatusText(s.Passes, isCurrentStory)

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

func (m *Model) mainScrollEnabled() bool {
	return m.phase != PhaseClarifying
}

func (m *Model) rebuildMainScrollContent() {
	if !m.mainScrollEnabled() {
		return
	}
	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteString("\n")
	b.WriteString(m.renderPhase())
	b.WriteString("\n")
	switch m.phase {
	case PhaseInit, PhasePRDGeneration:
		b.WriteString(m.renderGenerating())
	case PhaseFailed:
		b.WriteString(m.renderFailed())
	case PhasePRDReview:
		b.WriteString(m.renderPRDReview())
	case PhaseImplementation:
		b.WriteString(m.renderImplementation())
	case PhaseCompleted:
		b.WriteString(m.renderCompleted())
	}
	m.mainPane.SetContent(b.String())
}
