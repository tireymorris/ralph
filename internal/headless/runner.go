package headless

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"ralph/internal/clean"
	"ralph/internal/shared/config"
	"ralph/internal/shared/runner"
	"ralph/internal/shared/runpaths"
	"ralph/internal/shared/runstate"
	"ralph/internal/shared/session"
	webrunner "ralph/internal/web/runner"
	"ralph/internal/workflow"
	"ralph/internal/workflow/events"
)

type Runner struct {
	*session.Session
	cfg    *config.Config
	stderr io.Writer
	mu     sync.Mutex
}

func New(cfg *config.Config, r runner.RunnerInterface, stderr io.Writer) *Runner {
	s := session.NewWithRunner(cfg, r)
	s.SetReviewLoop(runstate.LocalRunID, workflow.NewFileReviewLoop(cfg.WorkDir, runstate.LocalRunID))
	if stderr == nil {
		stderr = os.Stderr
	}
	return &Runner{Session: s, cfg: cfg, stderr: stderr}
}

func Run(cfg *config.Config, prompt string, resume bool) int {
	return New(cfg, runner.New(cfg), os.Stderr).Run(prompt, resume)
}

func (r *Runner) Run(prompt string, resume bool) int {
	r.cfg.AutoApprove = true

	ctx := context.Background()
	if !resume {
		if _, err := clean.ArchivePriorState(r.cfg); err != nil {
			r.writeTerminalEvent(events.EventError{Err: fmt.Errorf("archive prior state: %w", err)})
			r.Wait()
			return 1
		}
		r.StartNew(ctx, prompt)
	} else {
		r.StartCheckpointResume(ctx)
	}

	return r.loop()
}

func (r *Runner) loop() int {
	for {
		ev, ok := <-r.EventsCh()
		if !ok {
			r.Wait()
			return 0
		}
		if err := r.writeEvent(ev); err != nil {
			r.Cancel()
			r.Wait()
			return 1
		}
		r.TrackEventState(ev)
		switch ev.(type) {
		case events.EventImplementationReview:
			go r.autoContinueImplementationReview()
		case events.EventCompleted:
			r.Wait()
			return 0
		case events.EventError:
			r.Wait()
			return 1
		}
	}
}

func (r *Runner) autoContinueImplementationReview() {
	time.Sleep(10 * time.Millisecond)
	if err := r.ContinueImplementationReview(context.Background(), r.cfg); err != nil {
		r.writeTerminalEvent(events.EventError{Err: err})
		r.Cancel()
	}
}

func (r *Runner) writeEvent(ev events.Event) error {
	data, err := webrunner.MarshalEventEnvelope(ev)
	if err != nil {
		return err
	}
	return r.writeLine(data)
}

func (r *Runner) writeTerminalEvent(ev events.Event) {
	_ = r.writeEvent(ev)
}

func (r *Runner) writeLine(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := appendRunEvent(r.cfg.WorkDir, runstate.LocalRunID, data); err != nil {
		return err
	}
	if _, err := r.stderr.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

func appendRunEvent(workDir, runID string, data []byte) error {
	path := runpaths.EventsPath(workDir, runID)
	if err := os.MkdirAll(runpaths.RunDir(workDir, runID), 0o750); err != nil {
		return fmt.Errorf("create run event dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open events log: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("append event: %w", err)
	}
	return nil
}
