package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/shared/config"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/prd"
	"ralph/internal/workflow"
	"ralph/internal/workflow/events"
)

type OperationManager struct {
	cfg        *config.Config
	ctx        context.Context
	cancelFunc context.CancelFunc
	eventsCh   chan events.Event
	executor   *workflow.Executor
}

func NewOperationManager(cfg *config.Config) *OperationManager {
	ctx, cancel := context.WithCancel(context.Background())
	eventsCh := make(chan events.Event, constants.EventChannelBuffer)

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

// StartFullOperation runs clarify then PRD generation in the background.
func (om *OperationManager) StartFullOperation(resume bool, userPrompt string) tea.Cmd {
	return func() tea.Msg {
		om.startBackground(func() {
			if resume {
				om.executor.RunLoad(om.ctx)
				return
			}
			qas, err := om.executor.RunClarify(om.ctx, userPrompt)
			if err != nil {
				om.emitError(clarifyPhaseError(err))
				return
			}
			om.executor.RunGenerateWithAnswers(om.ctx, userPrompt, qas)
		})

		// Return immediately so the UI shows PRD generation while work continues.
		return phaseChangeMsg(PhasePRDGeneration)
	}
}

func (om *OperationManager) StartImplementation(p *prd.PRD) tea.Cmd {
	return func() tea.Msg {
		om.startBackground(func() {
			om.executor.RunImplementation(om.ctx, p)
		})
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
	event events.Event
}
