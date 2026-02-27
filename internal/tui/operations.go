package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/config"
	"ralph/internal/constants"
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
	eventsCh := make(chan workflow.Event, constants.EventChannelBuffer)

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

// StartFullOperation launches the complete PRD workflow in a background
// goroutine: (1) clarifying questions (skipped on --resume), (2) PRD
// generation or load, (3) implementation. All results are communicated via the
// event channel and the returned tea.Cmd.
//
// For the clarifying phase the workflow blocks waiting for user answers, which
// arrive via EventClarifyingQuestions.AnswersCh — the TUI model sends them
// there when the user submits the form.
func (om *OperationManager) StartFullOperation(resume bool, userPrompt string) tea.Cmd {
	return func() tea.Msg {
		go func() {
			if resume {
				om.executor.RunLoad(om.ctx)
			} else {
				qas, err := om.executor.RunClarify(om.ctx, userPrompt)
				if err != nil {
					return
				}
				om.executor.RunGenerateWithAnswers(om.ctx, userPrompt, qas)
			}
		}()
		// Return a phase change immediately so the UI shows "Phase 1: PRD Generation"
		return phaseChangeMsg(PhasePRDGeneration)
	}
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
