package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runstate"
	"ralph/internal/workflow"
	"ralph/internal/workflow/events"
)

type OperationManager struct {
	*workflow.Driver
}

func NewOperationManager(cfg *config.Config) *OperationManager {
	d := workflow.NewDriver(cfg)
	d.SetReviewLoop(runstate.LocalRunID, workflow.NewFileReviewLoop(cfg.WorkDir, runstate.LocalRunID))
	return &OperationManager{Driver: d}
}

func (om *OperationManager) StartFullOperation(resume bool, userPrompt string) tea.Cmd {
	return func() tea.Msg {
		if resume {
			om.StartCheckpointResume(context.Background())
		} else {
			om.StartNew(context.Background(), userPrompt)
		}
		return phaseChangeMsg(PhasePRDGeneration)
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
