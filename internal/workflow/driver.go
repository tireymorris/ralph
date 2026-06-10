package workflow

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
	"ralph/internal/shared/runstate"
	"ralph/internal/workflow/events"
)

var ErrNotWaitingClarify = errors.New("run is not waiting for clarification")

type Driver struct {
	cfg      *config.Config
	ctx      context.Context
	cancel   context.CancelFunc
	eventsCh chan events.Event
	executor *Executor

	mu               sync.Mutex
	clarifyAnswersCh chan<- []prompt.QuestionAnswer
	currentPRD       *prd.PRD
	userPrompt       string
}

func NewDriver(cfg *config.Config) *Driver {
	ctx, cancel := context.WithCancel(context.Background())
	eventsCh := make(chan events.Event, constants.EventChannelBuffer)
	return &Driver{
		cfg:      cfg,
		ctx:      ctx,
		cancel:   cancel,
		eventsCh: eventsCh,
		executor: NewExecutor(cfg, eventsCh),
	}
}

func NewDriverWithRunner(cfg *config.Config, r runner.RunnerInterface) *Driver {
	ctx, cancel := context.WithCancel(context.Background())
	eventsCh := make(chan events.Event, constants.EventChannelBuffer)
	return &Driver{
		cfg:      cfg,
		ctx:      ctx,
		cancel:   cancel,
		eventsCh: eventsCh,
		executor: NewExecutorWithRunner(cfg, eventsCh, r),
	}
}

func (d *Driver) Ctx() context.Context          { return d.ctx }
func (d *Driver) EventsCh() <-chan events.Event { return d.eventsCh }
func (d *Driver) Cancel()                       { d.cancel() }

func (d *Driver) CurrentPRD() *prd.PRD {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.currentPRD
}

func (d *Driver) Prompt() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.userPrompt
}

func (d *Driver) StartNew(ctx context.Context, userPrompt string) {
	d.mu.Lock()
	d.userPrompt = userPrompt
	d.mu.Unlock()
	go d.runWithCtx(ctx, func(runCtx context.Context) {
		qas, err := d.executor.RunClarify(runCtx, userPrompt)
		if err != nil {
			d.EmitError(fmt.Errorf("clarification phase: %w", err))
			return
		}
		p, err := d.executor.RunGenerateWithAnswers(runCtx, userPrompt, qas)
		if err != nil || p == nil || !d.cfg.AutoApprove {
			return
		}
		d.executor.RunImplementation(runCtx, p)
	})
}

func (d *Driver) StartResume(ctx context.Context) {
	go d.runWithCtx(ctx, func(runCtx context.Context) {
		p, err := d.executor.RunLoad(runCtx)
		if err != nil || p == nil || !d.cfg.AutoApprove {
			return
		}
		d.executor.RunImplementation(runCtx, p)
	})
}

func (d *Driver) StartCheckpointResume(ctx context.Context) {
	go d.runWithCtx(ctx, func(runCtx context.Context) {
		checkpoint := d.reviewLoopCheckpoint()
		p, err := prd.Load(d.cfg)
		if err != nil {
			d.executor.RunLoad(runCtx)
			return
		}
		d.mu.Lock()
		d.currentPRD = p
		d.mu.Unlock()

		switch checkpoint {
		case runstate.CheckpointPRDReview:
			d.executor.RunLoad(runCtx)
		case runstate.CheckpointImplReview, runstate.CheckpointFollowup:
			d.executor.RunImplementation(runCtx, p)
		case runstate.CheckpointComplete:
			return
		default:
			if !p.AllCompleted() {
				d.executor.RunImplementation(runCtx, p)
			} else {
				d.executor.RunLoad(runCtx)
			}
		}
	})
}

func (d *Driver) reviewLoopCheckpoint() string {
	if r, ok := d.executor.reviewLoop.(checkpointReader); ok {
		return r.Checkpoint()
	}
	return ""
}

func (d *Driver) SetReviewLoop(runID string, updater ReviewLoopUpdater) {
	d.executor.SetReviewLoop(runID, updater)
}

func (d *Driver) StartImplementation(ctx context.Context, p *prd.PRD) {
	go d.runWithCtx(ctx, func(runCtx context.Context) {
		d.executor.RunImplementation(runCtx, p)
	})
}

func (d *Driver) ContinueImplementationReviewFromPRD(ctx context.Context, p *prd.PRD) {
	go d.runWithCtx(ctx, func(runCtx context.Context) {
		if err := d.executor.RunImplementationAfterReviewRecovery(runCtx, p); err != nil {
			d.EmitError(err)
		}
	})
}

func (d *Driver) StartCritiqueRevision(ctx context.Context, userPrompt, critique string) {
	go d.runWithCtx(ctx, func(runCtx context.Context) {
		if err := d.executor.RunCritiqueRevision(runCtx, userPrompt, critique); err != nil {
			d.EmitError(fmt.Errorf("critique revision: %w", err))
		}
	})
}

func (d *Driver) SubmitClarify(qas []prompt.QuestionAnswer) error {
	d.mu.Lock()
	ch := d.clarifyAnswersCh
	d.clarifyAnswersCh = nil
	d.mu.Unlock()
	if ch == nil {
		return ErrNotWaitingClarify
	}
	ch <- qas
	return nil
}

func (d *Driver) TrackEventState(ev events.Event) {
	d.mu.Lock()
	defer d.mu.Unlock()
	switch e := ev.(type) {
	case events.EventClarifyingQuestions:
		d.clarifyAnswersCh = e.AnswersCh
	case events.EventPRDGenerating, events.EventPRDRevising:
		d.clarifyAnswersCh = nil
	case events.EventPRDGenerated:
		d.currentPRD = e.PRD
	case events.EventPRDLoaded:
		d.currentPRD = e.PRD
	case events.EventPRDReview:
		d.currentPRD = e.PRD
	}
}

func (d *Driver) RunPrompt(ctx context.Context, p string, outputCh chan<- runner.OutputLine) error {
	return d.executor.RunPrompt(ctx, p, outputCh)
}

func (d *Driver) EmitError(err error) {
	if err == nil {
		return
	}
	select {
	case d.eventsCh <- events.EventError{Err: err}:
	case <-d.ctx.Done():
	}
}

func (d *Driver) EmitEvent(ev events.Event) {
	d.eventsCh <- ev
}

func (d *Driver) runWithCtx(parent context.Context, fn func(context.Context)) {
	runCtx, runCancel := context.WithCancel(parent)
	defer runCancel()

	go func() {
		select {
		case <-d.ctx.Done():
			runCancel()
		case <-runCtx.Done():
		}
	}()

	fn(runCtx)
}
