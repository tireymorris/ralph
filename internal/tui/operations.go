package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/clean"
	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runstate"
	"ralph/internal/workflow"
	"ralph/internal/workflow/events"
)

type OperationManager struct {
	*workflow.Driver
	cfg *config.Config
}

func NewOperationManager(cfg *config.Config) *OperationManager {
	d := workflow.NewDriver(cfg)
	d.SetReviewLoop(runstate.LocalRunID, workflow.NewFileReviewLoop(cfg.WorkDir, runstate.LocalRunID))
	return &OperationManager{Driver: d, cfg: cfg}
}

func (om *OperationManager) StartFullOperation(resume bool, userPrompt string) tea.Cmd {
	return func() tea.Msg {
		if resume {
			om.StartCheckpointResume(context.Background())
			return om.resumeStartMsg()
		}
		if _, err := clean.ArchivePriorState(om.cfg); err != nil {
			return operationErrorMsg{err: fmt.Errorf("archive prior state: %w", err)}
		}
		om.StartNew(context.Background(), userPrompt)
		return phaseChangeMsg(PhasePRDGeneration)
	}
}

func (om *OperationManager) resumeStartMsg() tea.Msg {
	p, err := prd.Load(om.cfg)
	if err != nil {
		return phaseChangeMsg(PhasePRDGeneration)
	}
	checkpoint := workflow.NewFileReviewLoop(om.cfg.WorkDir, runstate.LocalRunID).Checkpoint()
	return resumeStartMsg{phase: resumePhase(checkpoint, p), prd: p}
}

func resumePhase(checkpoint string, p *prd.PRD) Phase {
	switch checkpoint {
	case runstate.CheckpointPRDReview:
		return PhasePRDReview
	case runstate.CheckpointImplReview:
		return PhaseImplementationReview
	case runstate.CheckpointFollowup:
		return PhaseImplementation
	case runstate.CheckpointComplete:
		return PhaseCompleted
	default:
		if !p.AllCompleted() {
			return PhaseImplementation
		}
		return PhasePRDReview
	}
}

func (om *OperationManager) StartImplementation(p *prd.PRD) tea.Cmd {
	return func() tea.Msg {
		om.Driver.StartImplementation(context.Background(), p)
		return nil
	}
}

func (om *OperationManager) StartCritiqueRevision(userPrompt, critique string) tea.Cmd {
	return func() tea.Msg {
		om.Driver.StartCritiqueRevision(context.Background(), userPrompt, critique)
		return phaseChangeMsg(PhasePRDGeneration)
	}
}

func (om *OperationManager) ListenForEvents() tea.Cmd {
	return func() tea.Msg {
		select {
		case <-om.Ctx().Done():
			return nil
		case event, ok := <-om.EventsCh():
			if !ok {
				return nil
			}
			return workflowEventMsg{event: event}
		}
	}
}

type workflowEventMsg struct {
	event events.Event
}
