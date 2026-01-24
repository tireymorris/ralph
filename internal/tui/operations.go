package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/config"
	"ralph/internal/prd"
	"ralph/internal/workflow"
)

type OperationManager struct {
	cfg        *config.Config
	ctx        context.Context
	cancelFunc context.CancelFunc
	eventsCh   chan workflow.Event
	executor   *workflow.Executor
}

func NewOperationManager(cfg *config.Config) *OperationManager {
	ctx, cancel := context.WithCancel(context.Background())
	eventsCh := make(chan workflow.Event, 10000)

	return &OperationManager{
		cfg:        cfg,
		ctx:        ctx,
		cancelFunc: cancel,
		eventsCh:   eventsCh,
		executor:   workflow.NewExecutor(cfg, eventsCh),
	}
}

func (om *OperationManager) Cancel() {
	if om.cancelFunc != nil {
		om.cancelFunc()
	}
}

func (om *OperationManager) StartOperation() tea.Cmd {
	return func() tea.Msg {
		return phaseChangeMsg(PhasePRDGeneration)
	}
}

func (om *OperationManager) RunPRDOperation(resume bool, prompt string) tea.Msg {
	if resume {
		p, err := om.executor.RunLoad(om.ctx)
		if err != nil {
			return prdErrorMsg{err: err}
		}
		return prdGeneratedMsg{prd: p}
	}

	p, err := om.executor.RunGenerate(om.ctx, prompt)
	if err != nil {
		return prdErrorMsg{err: err}
	}
	return prdGeneratedMsg{prd: p}
}

func (om *OperationManager) StartImplementation(p *prd.PRD) tea.Cmd {
	return func() tea.Msg {
		go func() {
			om.executor.RunImplementation(om.ctx, p)
		}()
		return nil
	}
}

func (om *OperationManager) ListenForEvents() tea.Cmd {
	return func() tea.Msg {
		select {
		case <-om.ctx.Done():
			return nil
		case event, ok := <-om.eventsCh:
			if !ok {
				return nil
			}
			return workflowEventMsg{event: event}
		}
	}
}

type workflowEventMsg struct {
	event workflow.Event
}
