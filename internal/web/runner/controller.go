package runner

import (
	"context"
	"fmt"
	"os"
	"sync"

	"ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/runner"
	"ralph/internal/shared/runpaths"
	"ralph/internal/shared/runstate"
	"ralph/internal/shared/session"
	"ralph/internal/web/runs"
	"ralph/internal/workflow"
	"ralph/internal/workflow/events"
)

type RunController struct {
	*session.Session
	cfg       *config.Config
	registry  *runs.Registry
	lifecycle *runs.Lifecycle
	runID     string

	mu               sync.Mutex
	subscribers      map[chan events.Event]struct{}
	onTerminal       func()
	terminalNotified bool
	wg               sync.WaitGroup
}

func (c *RunController) SetOnTerminal(fn func()) {
	c.mu.Lock()
	c.onTerminal = fn
	c.mu.Unlock()
}

func (c *RunController) Wait() {
	c.Session.Wait()
	c.wg.Wait()
}

func NewControllerWithRunner(cfg *config.Config, registry *runs.Registry, runID string, r runner.RunnerInterface) *RunController {
	s := session.NewWithRunner(cfg, r)

	c := &RunController{
		Session:     s,
		cfg:         cfg,
		registry:    registry,
		lifecycle:   runs.NewLifecycle(registry),
		runID:       runID,
		subscribers: make(map[chan events.Event]struct{}),
	}
	s.SetReviewLoop(runID, newRegistryReviewLoop(registry, runID))
	c.wg.Add(1)
	go c.processEvents()
	return c
}

func (c *RunController) StartNew(ctx context.Context, userPrompt string) {
	c.Session.StartNew(ctx, userPrompt)
}

func (c *RunController) SubmitClarify(qas []prompt.QuestionAnswer) error {
	if err := c.Session.SubmitClarify(qas); err != nil {
		return err
	}
	_ = c.lifecycle.Clarify(c.runID)
	return nil
}

func (c *RunController) ContinueImplementationReview(ctx context.Context) error {
	if err := c.Session.ContinueImplementationReview(ctx, c.runConfig()); err != nil {
		return err
	}
	_ = c.lifecycle.ContinueImplementationReview(c.runID)
	return nil
}

func (c *RunController) ApproveReview(ctx context.Context) error {
	return c.Session.ApproveReview(ctx, c.runConfig())
}

func (c *RunController) ReviseReview(ctx context.Context, critique string) error {
	userPrompt := c.Prompt()
	if userPrompt == "" {
		if run, ok := c.registry.Get(c.runID); ok {
			userPrompt = run.Prompt
		}
	}
	if err := c.Session.ReviseReview(ctx, userPrompt, critique); err != nil {
		return err
	}
	_ = c.lifecycle.ReviseReview(c.runID)
	return nil
}

func (c *RunController) RunFollowUp(ctx context.Context, message string, cfg *config.Config) {
	_ = c.lifecycle.FollowUp(c.runID)
	fail := func(err error) {
		c.EmitEvent(events.EventError{Err: err})
	}

	run, ok := c.registry.Get(c.runID)
	if !ok {
		fail(fmt.Errorf("run %s not found", c.runID))
		return
	}

	transcript, err := runs.ReadEventTranscript(run.WorkDir, c.runID, 200)
	if err != nil {
		fail(fmt.Errorf("read run transcript: %w", err))
		return
	}

	runCfg := *cfg
	runCfg.WorkDir = run.WorkDir
	if run.PRDPath != "" {
		runCfg.PRDFile = run.PRDPath
	}

	if err := c.Session.RunFollowUp(ctx, &runCfg, message, transcript); err != nil {
		fail(err)
		return
	}
}

func (c *RunController) processEvents() {
	defer c.wg.Done()
	for {
		select {
		case <-c.Ctx().Done():
			return
		case ev, ok := <-c.EventsCh():
			if !ok {
				return
			}
			c.handleEvent(ev)
		}
	}
}

func (c *RunController) Subscribe() (<-chan events.Event, func()) {
	ch := make(chan events.Event, 64)
	c.mu.Lock()
	c.subscribers[ch] = struct{}{}
	c.mu.Unlock()
	unsub := func() {
		c.mu.Lock()
		delete(c.subscribers, ch)
		c.mu.Unlock()
	}
	return ch, unsub
}

func (c *RunController) fanOut(ev events.Event) {
	c.mu.Lock()
	subs := make([]chan events.Event, 0, len(c.subscribers))
	for ch := range c.subscribers {
		subs = append(subs, ch)
	}
	c.mu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- ev:
		default:
		}
	}
}

func (c *RunController) handleEvent(ev events.Event) {
	c.TrackEventState(ev)
	c.fanOut(ev)
	status, phase := workflow.EventStatusPhase(ev)
	if status != "" || phase != "" {
		if run, ok := c.registry.Get(c.runID); ok && run.Status == runstate.StatusCancelled {
			status, phase = "", ""
		}
	}
	if status != "" || phase != "" {
		_ = c.registry.UpdateStatus(c.runID, status, phase)
	}
	if checkpoint := workflow.EventCheckpoint(ev); checkpoint != "" {
		_ = c.registry.UpdateCheckpoint(c.runID, checkpoint)
	}
	_ = appendRunEvent(c.cfg.WorkDir, c.runID, ev)
	if runs.IsTerminalStatus(status) {
		c.mu.Lock()
		fn := c.onTerminal
		notify := fn != nil && !c.terminalNotified
		if notify {
			c.terminalNotified = true
		}
		c.mu.Unlock()
		if notify {
			fn()
		}
	}
}

func appendRunEvent(workDir, runID string, ev events.Event) error {
	line, err := marshalEventLine(ev)
	if err != nil {
		return err
	}
	path := runpaths.EventsPath(workDir, runID)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open events log: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("append event: %w", err)
	}
	return nil
}

func marshalEventLine(ev events.Event) ([]byte, error) {
	return MarshalEventEnvelope(ev)
}
