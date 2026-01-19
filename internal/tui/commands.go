package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/git"
	"ralph/internal/prd"
	"ralph/internal/runner"
	"ralph/internal/story"
)

func (m *Model) startOperation() tea.Cmd {
	return func() tea.Msg {
		if m.resume {
			return m.loadAndResume()
		}
		return m.generatePRD()
	}
}

func (m *Model) loadAndResume() tea.Msg {
	loadedPRD, err := prd.Load(m.cfg)
	if err != nil {
		return prdErrorMsg{err: err}
	}
	return prdGeneratedMsg{prd: loadedPRD}
}

func (m *Model) generatePRD() tea.Msg {
	gen := prd.NewGenerator(m.cfg)

	generatedPRD, err := gen.Generate(m.ctx, m.prompt, m.outputCh)
	if err != nil {
		return prdErrorMsg{err: err}
	}

	if err := prd.Save(m.cfg, generatedPRD); err != nil {
		return prdErrorMsg{err: fmt.Errorf("failed to save PRD: %w", err)}
	}

	return prdGeneratedMsg{prd: generatedPRD}
}

func (m *Model) setupBranchAndStart() tea.Cmd {
	return func() tea.Msg {
		if m.prd.BranchName != "" {
			gitMgr := git.New()
			if err := gitMgr.CreateBranch(m.prd.BranchName); err != nil {
				m.addLog(fmt.Sprintf("Warning: failed to create branch: %v", err))
			}
		}
		return m.startNextStory()
	}
}

func (m *Model) continueImplementation() tea.Cmd {
	return func() tea.Msg {
		if m.prd.AllCompleted() {
			prd.Delete(m.cfg)
			return phaseChangeMsg(PhaseCompleted)
		}

		next := m.prd.NextPendingStory(m.cfg.RetryAttempts)
		if next == nil {
			return phaseChangeMsg(PhaseFailed)
		}

		if m.iteration >= m.cfg.MaxIterations {
			return phaseChangeMsg(PhaseFailed)
		}

		return m.startNextStory()
	}
}

func (m *Model) startNextStory() tea.Msg {
	next := m.prd.NextPendingStory(m.cfg.RetryAttempts)
	if next == nil {
		if m.prd.AllCompleted() {
			return phaseChangeMsg(PhaseCompleted)
		}
		return phaseChangeMsg(PhaseFailed)
	}

	go func() {
		impl := story.NewImplementer(m.cfg)
		success, err := impl.Implement(m.ctx, next, m.iteration+1, m.prd, m.outputCh)

		if err != nil {
			m.outputCh <- runner.OutputLine{Text: fmt.Sprintf("Error: %v", err), IsErr: true}
		}

		if success {
			m.outputCh <- runner.OutputLine{Text: "STORY_COMPLETE:success"}
		} else {
			m.outputCh <- runner.OutputLine{Text: "STORY_COMPLETE:failure"}
		}
	}()

	return storyStartMsg{story: next}
}

func (m *Model) listenForOutput() tea.Cmd {
	return func() tea.Msg {
		select {
		case <-m.ctx.Done():
			return nil
		case line, ok := <-m.outputCh:
			if !ok {
				return nil
			}
			if strings.HasPrefix(line.Text, "STORY_COMPLETE:") {
				success := strings.HasSuffix(line.Text, "success")
				return storyCompleteMsg{success: success}
			}
			return outputMsg(line)
		}
	}
}
