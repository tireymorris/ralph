package session

import (
	"context"
	"fmt"

	"ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
	"ralph/internal/workflow"
	"ralph/internal/workflow/events"
)

// Session is a shared facade over workflow.Driver for UI-specific callers.
type Session struct {
	*workflow.Driver
}

func New(cfg *config.Config) *Session {
	return &Session{Driver: workflow.NewDriver(cfg)}
}

func NewWithRunner(cfg *config.Config, r runner.RunnerInterface) *Session {
	return &Session{Driver: workflow.NewDriverWithRunner(cfg, r)}
}

func (s *Session) ApproveReview(ctx context.Context, cfg *config.Config) error {
	p, err := s.PRDForImplementation(cfg)
	if err != nil {
		return err
	}
	s.StartImplementation(ctx, p)
	return nil
}

func (s *Session) ContinueImplementationReview(ctx context.Context, cfg *config.Config) error {
	p, err := s.PRDForImplementation(cfg)
	if err != nil {
		return err
	}
	s.ContinueImplementationReviewFromPRD(ctx, p)
	return nil
}

func (s *Session) StartImplementationFromPRD(ctx context.Context, p *prd.PRD) {
	s.StartImplementation(ctx, p)
}

func (s *Session) ContinueImplementationReviewFromPRD(ctx context.Context, p *prd.PRD) {
	s.Driver.ContinueImplementationReviewFromPRD(ctx, p)
}

func (s *Session) ResetPRDForImplementation(cfg *config.Config) (*prd.PRD, error) {
	p, err := prd.Load(cfg)
	if err != nil {
		return nil, fmt.Errorf("load PRD for implementation: %w", err)
	}
	p.UnmarkAllStories()
	if err := prd.Save(cfg, p); err != nil {
		return nil, fmt.Errorf("save PRD for implementation: %w", err)
	}
	p, err = prd.Load(cfg)
	if err != nil {
		return nil, fmt.Errorf("load PRD for implementation: %w", err)
	}
	return p, nil
}

func (s *Session) ReviseReview(ctx context.Context, userPrompt, critique string) error {
	s.StartCritiqueRevision(ctx, userPrompt, critique)
	return nil
}

func (s *Session) PRDForImplementation(cfg *config.Config) (*prd.PRD, error) {
	p := s.CurrentPRD()
	if p != nil {
		return p, nil
	}
	loaded, err := prd.Load(cfg)
	if err != nil {
		return nil, fmt.Errorf("load PRD for implementation: %w", err)
	}
	return loaded, nil
}

func (s *Session) SubmitClarify(qas []prompt.QuestionAnswer) error {
	return s.Driver.SubmitClarify(qas)
}
func (s *Session) EventsCh() <-chan events.Event { return s.Driver.EventsCh() }
func (s *Session) Wait()                         { s.Driver.Wait() }
