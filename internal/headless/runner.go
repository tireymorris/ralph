package headless

import (
	"context"
	"io"
	"os"
	"sync"

	"ralph/internal/shared/config"
	"ralph/internal/shared/runner"
	"ralph/internal/shared/runstate"
	"ralph/internal/shared/session"
	"ralph/internal/workflow/events"
)

type Runner struct {
	*session.Session
	cfg      *config.Config
	stderr   io.Writer
	mu       sync.Mutex
	snapshot session.RunSnapshot
}

func New(cfg *config.Config, r runner.RunnerInterface, stderr io.Writer) *Runner {
	if stderr == nil {
		stderr = os.Stderr
	}
	return &Runner{
		Session: session.NewWithRunner(cfg, r),
		cfg:     cfg,
		stderr:  stderr,
	}
}

func Run(cfg *config.Config, prompt string, resume bool) int {
	return New(cfg, runner.New(cfg), os.Stderr).Run(prompt, resume)
}

func (r *Runner) Run(prompt string, resume bool) int {
	r.cfg.AutoApprove = true

	opts := session.UnattendedOptions{Prompt: prompt, Resume: resume}
	if err := r.StartUnattended(context.Background(), r.cfg, opts); err != nil {
		_ = r.writeTerminalEvent(events.EventError{Err: err})
		r.Wait()
		return 1
	}

	sink := newNDJSONSink(r.cfg.WorkDir, runstate.LocalRunID, r.stderr, r.refreshSnapshot)
	return r.RunEventLoop(sink)
}

func (r *Runner) Snapshot() session.RunSnapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.snapshot
}

func (r *Runner) writeTerminalEvent(ev events.Event) error {
	sink := newNDJSONSink(r.cfg.WorkDir, runstate.LocalRunID, r.stderr, nil)
	_, _, err := sink.OnEvent(ev)
	return err
}

func (r *Runner) refreshSnapshot() {
	snapshot := r.RunSnapshot(runstate.PhaseImplement)
	r.mu.Lock()
	r.snapshot = snapshot
	r.mu.Unlock()
}
