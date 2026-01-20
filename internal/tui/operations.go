package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal"
	"ralph/internal/config"
	"ralph/internal/git"
	"ralph/internal/prd"
	"ralph/internal/runner"
)

// OperationManager handles all operations like PRD generation and story implementation
type OperationManager struct {
	cfg         *config.Config
	ctx         context.Context
	cancelFunc  context.CancelFunc
	outputCh    chan runner.OutputLine
	generator   internal.PRDGenerator
	implementer internal.StoryImplementer
}

// NewOperationManager creates a new operation manager
func NewOperationManager(cfg *config.Config, generator internal.PRDGenerator, implementer internal.StoryImplementer) *OperationManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &OperationManager{
		cfg:         cfg,
		ctx:         ctx,
		cancelFunc:  cancel,
		outputCh:    make(chan runner.OutputLine, 10000), // Large buffer to handle high-volume output
		generator:   generator,
		implementer: implementer,
	}
}

// GetContext returns the operation context
func (om *OperationManager) GetContext() context.Context {
	return om.ctx
}

// GetOutputChannel returns the output channel
func (om *OperationManager) GetOutputChannel() chan runner.OutputLine {
	return om.outputCh
}

// Cancel cancels all operations
func (om *OperationManager) Cancel() {
	if om.cancelFunc != nil {
		om.cancelFunc()
	}
}

// StartOperation starts the initial operation (PRD generation)
func (om *OperationManager) StartOperation() tea.Cmd {
	// First, change the phase to show we're generating/loading
	return func() tea.Msg {
		return phaseChangeMsg(PhasePRDGeneration)
	}
}

// RunPRDOperation runs the PRD generation or loading operation
func (om *OperationManager) RunPRDOperation(resume bool, prompt string) tea.Msg {
	if resume {
		return om.loadAndResume()
	}
	return om.generatePRD(prompt)
}

// loadAndResume loads existing PRD for resume operation
func (om *OperationManager) loadAndResume() tea.Msg {
	// Send feedback
	if om.outputCh != nil {
		om.outputCh <- runner.OutputLine{Text: "Loading existing PRD..."}
	}

	loadedPRD, err := prd.Load(om.cfg)
	if err != nil {
		return prdErrorMsg{err: err}
	}
	return prdGeneratedMsg{prd: loadedPRD}
}

// generatePRD generates a new PRD
func (om *OperationManager) generatePRD(prompt string) tea.Msg {
	// Send initial feedback
	if om.outputCh != nil {
		om.outputCh <- runner.OutputLine{Text: "Analyzing codebase and generating PRD..."}
	}

	generatedPRD, err := om.generator.Generate(om.ctx, prompt, om.outputCh)
	if err != nil {
		return prdErrorMsg{err: err}
	}

	if err := prd.Save(om.cfg, generatedPRD); err != nil {
		return prdErrorMsg{err: fmt.Errorf("failed to save PRD: %w", err)}
	}

	return prdGeneratedMsg{prd: generatedPRD}
}

// SetupBranchAndStart sets up git branch and starts implementation
func (om *OperationManager) SetupBranchAndStart(branchName string, p *prd.PRD) tea.Cmd {
	// Capture values to avoid race conditions
	workDir := om.cfg.WorkDir

	return func() tea.Msg {
		if branchName != "" {
			gitMgr := git.NewWithWorkDir(workDir)
			if err := gitMgr.CreateBranch(branchName); err != nil {
				// Send warning through channel
				om.outputCh <- runner.OutputLine{Text: fmt.Sprintf("Warning: failed to create branch: %v", err), IsErr: true}
			}
		}
		return om.startNextStory(p, 1) // iteration starts at 1
	}
}

// ContinueImplementation continues with the next story or completes
func (om *OperationManager) ContinueImplementation(p *prd.PRD, iteration int) tea.Cmd {
	return func() tea.Msg {
		if p.AllCompleted() {
			prd.Delete(om.cfg)
			return phaseChangeMsg(PhaseCompleted)
		}

		next := p.NextPendingStory(om.cfg.RetryAttempts)
		if next == nil {
			return phaseChangeMsg(PhaseFailed)
		}

		if iteration >= om.cfg.MaxIterations {
			return phaseChangeMsg(PhaseFailed)
		}

		return om.startNextStory(p, iteration+1)
	}
}

// startNextStory starts the next pending story
func (om *OperationManager) startNextStory(p *prd.PRD, iteration int) tea.Msg {
	var next *prd.Story
	if p != nil {
		next = p.NextPendingStory(om.cfg.RetryAttempts)
	}

	if next == nil {
		if p != nil && p.AllCompleted() {
			return phaseChangeMsg(PhaseCompleted)
		}
		return phaseChangeMsg(PhaseFailed)
	}

	// Capture values to avoid race conditions
	ctx := om.ctx
	outputCh := om.outputCh
	prdCopy := p
	implementer := om.implementer

	go func(story *prd.Story, iter int, prd *prd.PRD, ch chan<- runner.OutputLine) {
		success, err := implementer.Implement(ctx, story, iter, prd, ch)

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

// ListenForOutput listens for output from operations
func (om *OperationManager) ListenForOutput() tea.Cmd {
	return func() tea.Msg {
		select {
		case <-om.ctx.Done():
			return nil
		case line, ok := <-om.outputCh:
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
