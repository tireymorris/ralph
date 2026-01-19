package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/git"
	"ralph/internal/prd"
	"ralph/internal/runner"
)

type PRDGenerator interface {
	Generate(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) (*prd.PRD, error)
}

type StoryImplementer interface {
	Implement(ctx context.Context, s *prd.Story, iteration int, p *prd.PRD, outputCh chan<- runner.OutputLine) (bool, error)
}

func (m *Model) startOperation() tea.Cmd {
	// First, change the phase to show we're generating/loading
	return func() tea.Msg {
		return phaseChangeMsg(PhasePRDGeneration)
	}
}

func (m *Model) runPRDOperation() tea.Cmd {
	return func() tea.Msg {
		if m.resume {
			return m.loadAndResume()
		}
		return m.generatePRD()
	}
}

func (m *Model) loadAndResume() tea.Msg {
	// Send feedback
	if m.outputCh != nil {
		m.outputCh <- runner.OutputLine{Text: "Loading existing PRD..."}
	}

	loadedPRD, err := prd.Load(m.cfg)
	if err != nil {
		return prdErrorMsg{err: err}
	}
	return prdGeneratedMsg{prd: loadedPRD}
}

func (m *Model) generatePRD() tea.Msg {
	// Send initial feedback
	if m.outputCh != nil {
		m.outputCh <- runner.OutputLine{Text: "Analyzing codebase and generating PRD..."}
	}

	generatedPRD, err := m.generator.Generate(m.ctx, m.prompt, m.outputCh)
	if err != nil {
		return prdErrorMsg{err: err}
	}

	if err := prd.Save(m.cfg, generatedPRD); err != nil {
		return prdErrorMsg{err: fmt.Errorf("failed to save PRD: %w", err)}
	}

	return prdGeneratedMsg{prd: generatedPRD}
}

func (m *Model) setupBranchAndStart() tea.Cmd {
	// Capture values to avoid race conditions
	branchName := m.prd.BranchName
	workDir := m.cfg.WorkDir
	outputCh := m.outputCh

	return func() tea.Msg {
		if branchName != "" {
			gitMgr := git.NewWithWorkDir(workDir)
			if err := gitMgr.CreateBranch(branchName); err != nil {
				// Send warning through channel instead of calling m.addLog directly
				outputCh <- runner.OutputLine{Text: fmt.Sprintf("Warning: failed to create branch: %v", err), IsErr: true}
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

	// Capture values to avoid race conditions - these are passed explicitly to the goroutine
	iteration := m.iteration + 1
	ctx := m.ctx
	outputCh := m.outputCh
	prdCopy := m.prd
	implementer := m.implementer

	go func(story *prd.Story, iter int, p *prd.PRD, ch chan<- runner.OutputLine) {
		success, err := implementer.Implement(ctx, story, iter, p, ch)

		if err != nil {
			ch <- runner.OutputLine{Text: fmt.Sprintf("Error: %v", err), IsErr: true}
		}

		if success {
			ch <- runner.OutputLine{Text: "STORY_COMPLETE:success"}
		} else {
			ch <- runner.OutputLine{Text: "STORY_COMPLETE:failure"}
		}
	}(next, iteration, prdCopy, outputCh)

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
