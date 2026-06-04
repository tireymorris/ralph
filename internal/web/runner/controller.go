package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
	"ralph/internal/web/runs"
	"ralph/internal/workflow"
	"ralph/internal/workflow/events"
)

var errNoPRDForImplement = errors.New("no PRD available for implementation")

type RunController struct {
	*workflow.Driver
	cfg      *config.Config
	registry *runs.Registry
	runID    string

	mu          sync.Mutex
	subscribers map[chan events.Event]struct{}
	onTerminal  func()
}

func (c *RunController) SetOnTerminal(fn func()) {
	c.mu.Lock()
	c.onTerminal = fn
	c.mu.Unlock()
}

func NewControllerWithRunner(cfg *config.Config, registry *runs.Registry, runID string, r runner.RunnerInterface) *RunController {
	d := workflow.NewDriverWithRunner(cfg, r)

	c := &RunController{
		Driver:      d,
		cfg:         cfg,
		registry:    registry,
		runID:       runID,
		subscribers: make(map[chan events.Event]struct{}),
	}
	go c.processEvents()
	return c
}

func (c *RunController) StartNew(ctx context.Context, userPrompt string) {
	c.Driver.StartNew(ctx, userPrompt)
}

func (c *RunController) SubmitClarify(qas []prompt.QuestionAnswer) error {
	if err := c.Driver.SubmitClarify(qas); err != nil {
		return err
	}
	_ = c.registry.UpdateStatus(c.runID, "running", "generate")
	return nil
}

func (c *RunController) ApproveReview(ctx context.Context) error {
	p := c.CurrentPRD()
	if p == nil {
		runCfg := c.runConfig()
		loaded, err := prd.Load(runCfg)
		if err != nil {
			return fmt.Errorf("load PRD for implementation: %w", err)
		}
		p = loaded
	}
	c.Driver.StartImplementation(ctx, p)
	return nil
}

func (c *RunController) ReviseReview(ctx context.Context, critique string) error {
	userPrompt := c.Prompt()
	if userPrompt == "" {
		if run, ok := c.registry.Get(c.runID); ok {
			userPrompt = run.Prompt
		}
	}
	c.Driver.StartCritiqueRevision(ctx, userPrompt, critique)
	_ = c.registry.UpdateStatus(c.runID, "running", "generate")
	return nil
}

func (c *RunController) RunFollowUp(ctx context.Context, message string, cfg *config.Config) {
	_ = c.registry.UpdateStatus(c.runID, "running", "followup")
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

	revisionPrompt := prompt.FollowUpRevision(message, runCfg.PRDFile, transcript)
	outputCh := make(chan runner.OutputLine, constants.EventChannelBuffer)
	done := make(chan struct{})
	go func() {
		workflow.NewOutputForwarder(c.EmitEvent).Forward(outputCh)
		close(done)
	}()
	if err := c.RunPrompt(ctx, revisionPrompt, outputCh); err != nil {
		close(outputCh)
		<-done
		fail(fmt.Errorf("follow-up revision: %w", err))
		return
	}
	close(outputCh)
	<-done

	p, err := prd.Load(&runCfg)
	if err != nil {
		fail(fmt.Errorf("load PRD: %w", err))
		return
	}
	p.UnmarkAllStories()
	if err := prd.Save(&runCfg, p); err != nil {
		fail(fmt.Errorf("save PRD: %w", err))
		return
	}

	p, err = prd.Load(&runCfg)
	if err != nil {
		fail(fmt.Errorf("reload PRD: %w", err))
		return
	}

	c.Driver.StartImplementation(ctx, p)
}

func (c *RunController) processEvents() {
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
	status, phase := mapEventToStatusPhase(ev)
	if status != "" || phase != "" {
		if run, ok := c.registry.Get(c.runID); ok && run.Status == "cancelled" {
			status, phase = "", ""
		}
	}
	if status != "" || phase != "" {
		_ = c.registry.UpdateStatus(c.runID, status, phase)
	}
	_ = appendRunEvent(c.cfg.WorkDir, c.runID, ev)
	if runs.IsTerminalStatus(status) {
		c.mu.Lock()
		fn := c.onTerminal
		c.mu.Unlock()
		if fn != nil {
			fn()
		}
	}
}

func mapEventToStatusPhase(ev events.Event) (status, phase string) {
	switch ev.(type) {
	case events.EventClarifyingQuestions:
		return "waiting_clarify", "clarify"
	case events.EventPRDGenerating, events.EventPRDRevising:
		return "running", "generate"
	case events.EventPRDGenerated, events.EventPRDLoaded, events.EventPRDReview:
		return "waiting_review", "review"
	case events.EventStoryStarted, events.EventStoryCompleted:
		return "implementing", "implement"
	case events.EventCleanupStarted, events.EventCleanupCompleted:
		return "implementing", "cleanup"
	case events.EventCompleted:
		return "completed", "complete"
	case events.EventError:
		return "failed", "failed"
	default:
		return "", ""
	}
}

func appendRunEvent(workDir, runID string, ev events.Event) error {
	line, err := marshalEventLine(ev)
	if err != nil {
		return err
	}
	path := filepath.Join(workDir, ".ralph", "runs", runID, "events.ndjson")
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
