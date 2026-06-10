package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"ralph/internal/shared/prd"
)

func (m *Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var b strings.Builder

	if m.phase == PhaseClarifying {
		b.WriteString(m.renderClarifyingView())
		return b.String()
	}

	if m.phase == PhaseAwaitingPrompt {
		b.WriteString(m.renderAwaitingPromptView())
		return b.String()
	}

	if m.scrollPane == focusMain {
		b.WriteString(m.mainPane.View())
	} else {
		b.WriteString(m.renderLogsPane())
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render(m.helpText()))

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

	contentWidth := 76
	if m.width > 0 {
		contentWidth = max(20, m.width-4)
	}

	var b strings.Builder

	instruction := "Please answer the following questions before we generate your PRD."
	b.WriteString(infoStyle.Render(inProgressStyle.Render(wrapText(instruction, contentWidth))))
	b.WriteString("\n")
	navHint := "  Tab/↑/↓ to navigate  •  Enter to confirm  •  Esc to skip all questions"
	b.WriteString(mutedStyle.Render(wrapText(navHint, contentWidth)))
	b.WriteString("\n\n")

	for i, q := range m.clarifyQuestions {
		num := labelStyle.Render(fmt.Sprintf("Q%d.", i+1))
		wrapped := wrapText(q, contentWidth)
		lines := strings.Split(wrapped, "\n")
		b.WriteString(infoStyle.Render(fmt.Sprintf("%s %s", num, valueStyle.Render(lines[0]))))
		for _, line := range lines[1:] {
			b.WriteString("\n")
			b.WriteString(infoStyle.Render(valueStyle.Render(line)))
		}
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
		wrapWidth := max(20, m.mainPane.Width-10)
		msg = wrapText(m.err.Error(), wrapWidth)
	}
	return errorStyle.Render(msg)
}

func (m *Model) renderGenerating() string {
	promptLabel := labelStyle.Render("Prompt")
	promptTextStyle := lipgloss.NewStyle().Foreground(textColor)
	wrapWidth := max(20, m.mainPane.Width-10)
	promptText := promptTextStyle.Render(wrapText(m.prompt, wrapWidth))
	generatingText := inProgressStyle.Render("Generating PRD from your requirements...")
	if m.revisingPRD {
		generatingText = inProgressStyle.Render("Revising PRD based on your critique...")
	}

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
	b.WriteString(m.renderProjectSection())
	b.WriteString("\n\n")
	b.WriteString(titleStyle.Render("Stories"))
	b.WriteString("\n")
	for _, s := range m.prd.Stories {
		b.WriteString(m.renderReviewStory(s))
	}
	b.WriteString("\n")
	if m.critiqueActive {
		b.WriteString(helpStyle.Render("Critique (Enter submit • Esc cancel)"))
		b.WriteString("\n")
		b.WriteString(infoStyle.Render(storyItemStyle.Render(m.critiqueInput.View())))
		b.WriteString("\n")
	}
	b.WriteString(helpStyle.Render("Press c to add critique or Enter to continue to implementation"))

	return b.String()
}

func (m *Model) renderImplementation() string {
	if m.prd == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString(m.renderProjectSection())
	b.WriteString("\n")
	b.WriteString(m.renderProgressSection())
	b.WriteString("\n\n")
	b.WriteString(titleStyle.Render("Stories"))
	b.WriteString("\n")
	for _, s := range m.prd.Stories {
		b.WriteString(m.renderImplementationStory(s))
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

func (m *Model) renderAwaitingPromptView() string {
	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteString("\n")
	b.WriteString(m.renderPhase())
	b.WriteString("\n")
	input := m.promptInput
	width := m.width
	if width <= 0 {
		width = 80
	}
	input.Width = max(20, width-4)
	b.WriteString(input.View())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(m.helpText()))
	return b.String()
}

func (m *Model) renderClarifyingView() string {
	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteString("\n")
	b.WriteString(m.renderPhase())
	b.WriteString("\n")
	b.WriteString(m.renderClarifying())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(m.clarifyingHelpText()))
	return b.String()
}

func (m *Model) renderLogsPane() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Output Logs"))
	b.WriteString("\n")
	b.WriteString(m.renderLogs())
	return b.String()
}

func (m *Model) helpText() string {
	if m.phase == PhaseAwaitingPrompt {
		return "enter: submit  q/ctrl+c: quit"
	}
	if m.phase == PhasePRDReview {
		return "Tab switch pane • ↑/↓ scroll • c critique • Enter continue • q quit • ctrl+c exit"
	}
	if m.phase == PhaseFailed {
		return "Tab switch pane • ↑/↓ scroll • r retry • q quit • ctrl+c exit"
	}
	return "Tab switch pane • ↑/↓ scroll • q quit • ctrl+c exit"
}

func (m *Model) clarifyingHelpText() string {
	return "Tab/↑/↓ navigate • Enter confirm • Esc skip all • ctrl+c exit"
}

func (m *Model) mainScrollEnabled() bool {
	return m.phase != PhaseClarifying && m.phase != PhaseAwaitingPrompt
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
	case PhaseImplementationReview, PhaseImplementation:
		b.WriteString(m.renderImplementation())
	case PhaseCompleted:
		b.WriteString(m.renderCompleted())
	}
	m.mainPane.SetContent(b.String())
}

func (m *Model) renderProjectSection() string {
	var b strings.Builder
	b.WriteString(m.renderProjectLine())
	if m.prd.BranchName != "" {
		b.WriteString("\n")
		b.WriteString(m.renderBranchLine())
	}
	return b.String()
}

func (m *Model) renderProjectLine() string {
	return infoStyle.Render(labelStyle.Render("Project") + " " + valueStyle.Render(m.prd.ProjectName))
}

func (m *Model) renderBranchLine() string {
	return infoStyle.Render(labelStyle.Render("Branch") + " " + valueStyle.Render(m.prd.BranchName))
}

func (m *Model) renderProgressSection() string {
	completed := m.prd.CompletedCount()
	total := len(m.prd.Stories)
	percent := float64(completed) / float64(total)
	var b strings.Builder
	b.WriteString(infoStyle.Render(labelStyle.Render("Progress") + " " + mutedStyle.Render(fmt.Sprintf("%d/%d stories", completed, total))))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render(m.progress.ViewAs(percent)))
	return b.String()
}

func (m *Model) renderReviewStory(s *prd.Story) string {
	var b strings.Builder
	status := "[ ]"
	if s.Passes {
		status = "[x]"
	}
	deps := ""
	if len(s.DependsOn) > 0 {
		deps = " (depends: " + strings.Join(s.DependsOn, ", ") + ")"
	}
	b.WriteString(storyItemStyle.Render(fmt.Sprintf("%s P%d %s%s", status, s.Priority, s.Title, deps)))
	b.WriteString("\n")
	if len(s.AcceptanceCriteria) > 0 {
		b.WriteString(mutedStyle.Render("    Acceptance criteria:"))
		b.WriteString("\n")
		wrapWidth := max(20, m.mainPane.Width-6)
		for _, ac := range s.AcceptanceCriteria {
			wrapped := wrapText(ac, wrapWidth)
			lines := strings.Split(wrapped, "\n")
			for i, line := range lines {
				if i == 0 {
					b.WriteString(mutedStyle.Render("      - " + line))
				} else {
					b.WriteString(mutedStyle.Render("        " + line))
				}
				b.WriteString("\n")
			}
		}
	}
	return b.String()
}

func (m *Model) renderImplementationStory(s *prd.Story) string {
	isCurrentStory := m.currentStory != nil && s.ID == m.currentStory.ID
	icon := getStatusIcon(s.Passes, isCurrentStory)
	status := getStatusText(s.Passes, isCurrentStory)
	line := fmt.Sprintf("%s %s  %s", icon, s.Title, status)
	if isCurrentStory {
		return selectedStoryStyle.Render(line)
	}
	return storyItemStyle.Render(line)
}
