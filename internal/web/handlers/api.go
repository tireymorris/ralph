package handlers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/logger"
	sharedrunner "ralph/internal/shared/runner"
	runctrl "ralph/internal/web/runner"
	"ralph/internal/web/runs"
)

type runnerFactory func(*config.Config) (sharedrunner.RunnerInterface, error)

type controllerFactory func(*config.Config, *runs.Registry, string, sharedrunner.RunnerInterface) *runctrl.RunController

type API struct {
	cfg               *config.Config
	registry          *runs.Registry
	runnerFactory     runnerFactory
	controllerFactory controllerFactory
	mu                sync.Mutex
	controllers       map[string]*runctrl.RunController
	releaseMu         sync.Mutex
	releaseWG         sync.WaitGroup
}

func NewAPI(cfg *config.Config, registry *runs.Registry) *API {
	return &API{
		cfg:      cfg,
		registry: registry,
		runnerFactory: func(c *config.Config) (sharedrunner.RunnerInterface, error) {
			return sharedrunner.NewWithError(c)
		},
		controllerFactory: runctrl.NewControllerWithRunner,
		controllers:       make(map[string]*runctrl.RunController),
	}
}

func (a *API) Cfg() *config.Config {
	return a.cfg
}

func (a *API) SetRunnerFactory(f runnerFactory) {
	a.runnerFactory = f
}

func (a *API) SetController(id string, ctrl *runctrl.RunController) {
	a.registerController(id, ctrl)
}

func (a *API) controller(id string) (*runctrl.RunController, bool) {
	a.mu.Lock()
	ctrl, ok := a.controllers[id]
	a.mu.Unlock()
	return ctrl, ok
}

func (a *API) registerController(id string, ctrl *runctrl.RunController) {
	a.mu.Lock()
	a.registerControllerLocked(id, ctrl)
	a.mu.Unlock()
}

func (a *API) releaseController(id string) {
	a.mu.Lock()
	ctrl := a.controllers[id]
	a.mu.Unlock()
	if ctrl != nil {
		ctrl.Cancel()
		ctrl.Wait()
	}
	a.mu.Lock()
	delete(a.controllers, id)
	a.mu.Unlock()
}

func (a *API) runConfigFor(run *runs.Run) *config.Config {
	runCfg := *a.cfg
	runCfg.WorkDir = run.WorkDir
	if run.PRDPath != "" {
		runCfg.PRDFile = run.PRDPath
	}
	runCfg.AutoApprove = run.AutoApprove
	return &runCfg
}

func (a *API) ensureController(id string) (*runctrl.RunController, error) {
	run, ok := a.registry.Get(id)
	if !ok {
		return nil, fmt.Errorf("run %q not found", id)
	}

	a.mu.Lock()
	if ctrl := a.controllers[id]; ctrl != nil {
		a.mu.Unlock()
		return ctrl, nil
	}
	a.mu.Unlock()

	runCfg := a.runConfigFor(run)
	rn, err := a.runnerFactory(runCfg)
	if err != nil {
		return nil, err
	}
	ctrl := a.controllerFactory(runCfg, a.registry, id, rn)

	a.mu.Lock()
	defer a.mu.Unlock()
	if existing := a.controllers[id]; existing != nil {
		ctrl.Cancel()
		return existing, nil
	}
	a.registerControllerLocked(id, ctrl)
	if runs.NeedsSessionReattach(run.Status) {
		ctrl.Reattach(context.Background())
	}
	return ctrl, nil
}

// ReattachInterruptedRuns restores workflow sessions for runs left non-terminal on disk.
func (a *API) ReattachInterruptedRuns() {
	for _, run := range a.registry.List() {
		if runs.IsTerminalStatus(run.Status) || !runs.NeedsSessionReattach(run.Status) {
			continue
		}
		if _, err := a.ensureController(run.ID); err != nil {
			logger.Warn("reattach interrupted run", "id", run.ID, "error", err)
		}
	}
}

func (a *API) registerControllerLocked(id string, ctrl *runctrl.RunController) {
	ctrl.SetOnTerminal(func() {
		a.releaseMu.Lock()
		a.releaseWG.Add(1)
		a.releaseMu.Unlock()
		go func() {
			defer a.releaseWG.Done()
			a.releaseController(id)
		}()
	})
	a.controllers[id] = ctrl
}

func (a *API) ReleaseAllControllers() {
	a.mu.Lock()
	ids := make([]string, 0, len(a.controllers))
	for id := range a.controllers {
		ids = append(ids, id)
	}
	a.mu.Unlock()
	for _, id := range ids {
		a.releaseController(id)
	}
	a.releaseMu.Lock()
	a.releaseWG.Wait()
	a.releaseMu.Unlock()
	time.Sleep(25 * time.Millisecond)
}
