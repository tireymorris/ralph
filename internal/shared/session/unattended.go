package session

import (
	"context"
	"fmt"

	"ralph/internal/clean"
	"ralph/internal/shared/config"
	"ralph/internal/shared/runstate"
	"ralph/internal/workflow"
	"ralph/internal/workflow/events"
)

type UnattendedOptions struct {
	Prompt string
	Resume bool
}

type EventSink interface {
	OnEvent(ev events.Event) (stop bool, exitCode int, err error)
}

func ConfigureLocalReviewLoop(cfg *config.Config, s *Session) {
	if s == nil || cfg == nil {
		return
	}
	s.SetReviewLoop(runstate.LocalRunID, workflow.NewFileReviewLoop(cfg.WorkDir, runstate.LocalRunID))
}

func (s *Session) StartUnattended(ctx context.Context, cfg *config.Config, opts UnattendedOptions) error {
	ConfigureLocalReviewLoop(cfg, s)
	if !opts.Resume {
		if _, err := clean.ArchivePriorState(cfg); err != nil {
			return fmt.Errorf("archive prior state: %w", err)
		}
		s.StartNew(ctx, opts.Prompt)
		return nil
	}
	s.StartCheckpointResume(ctx)
	return nil
}

func (s *Session) RunEventLoop(sink EventSink) int {
	if s == nil || sink == nil {
		return 1
	}
	for {
		ev, ok := <-s.EventsCh()
		if !ok {
			s.Wait()
			return 0
		}
		s.TrackEventState(ev)
		stop, code, err := sink.OnEvent(ev)
		if err != nil {
			s.Cancel()
			s.Wait()
			return 1
		}
		if stop {
			s.Wait()
			return code
		}
	}
}
