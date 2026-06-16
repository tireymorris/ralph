package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/clean"
	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runstate"
	"ralph/internal/shared/session"
	"ralph/internal/workflow"
	"ralph/internal/workflow/events"
)

type OperationManager struct {
	*session.Session
	cfg *config.Config
}

func NewOperationManager(cfg *config.Config) *OperationManager {
	s := session.New(cfg)
	s.SetReviewLoop(runstate.LocalRunID, workflow.NewFileReviewLoop(cfg.WorkDir, runstate.LocalRunID))
	return &OperationManager{Session: s, cfg: cfg}
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
	snapshot := session.RunSnapshot{
		Prompt:           om.Session.Prompt(),
		Phase:            runstate.CheckpointPhase(checkpoint, p),
		CurrentPRD:       p,
		CurrentStory:     p.NextReadyStory(),
		CompletedStories: p.CompletedCount(),
		TotalStories:     len(p.Stories),
	}
	if snapshot.CurrentStory != nil {
		snapshot.NextPendingSlice = snapshot.CurrentStory.NextPendingSlice()
	}
	return resumeStartMsg{phase: resumePhase(checkpoint, p), prd: p, snapshot: snapshot}
}

func resumePhase(checkpoint string, p *prd.PRD) Phase {
	switch runstate.CheckpointPhase(checkpoint, p) {
	case runstate.PhaseReview:
		return PhasePRDReview
	case runstate.PhaseImplementationReview:
		return PhaseImplementationReview
	case runstate.PhaseFollowup, runstate.PhaseImplement:
		return PhaseImplementation
	case runstate.PhaseCompleted:
		return PhaseCompleted
	default:
		return PhasePRDReview
	}
}

func (om *OperationManager) StartImplementation(p *prd.PRD) tea.Cmd {
	return func() tea.Msg {
		if p == nil {
			var err error
			p, err = om.PRDForImplementation(om.cfg)
			if err != nil {
				return operationErrorMsg{err: err}
			}
		}
		om.StartImplementationFromPRD(context.Background(), p)
		return nil
	}
}

func (om *OperationManager) ApproveReview() tea.Cmd {
	return func() tea.Msg {
		if err := om.Session.ApproveReview(context.Background(), om.cfg); err != nil {
			return operationErrorMsg{err: err}
		}
		return nil
	}
}

func (om *OperationManager) ContinueImplementationReview() tea.Cmd {
	return func() tea.Msg {
		if err := om.Session.ContinueImplementationReview(context.Background(), om.cfg); err != nil {
			return operationErrorMsg{err: err}
		}
		return nil
	}
}

func (om *OperationManager) StartCritiqueRevision(userPrompt, critique string) tea.Cmd {
	return func() tea.Msg {
		if err := om.ReviseReview(context.Background(), userPrompt, critique); err != nil {
			return operationErrorMsg{err: err}
		}
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
