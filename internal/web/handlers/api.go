package handlers

import (
	"sync"

	"ralph/internal/shared/config"
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

func (a *API) registerController(id string, ctrl *runctrl.RunController) {
	a.mu.Lock()
	a.registerControllerLocked(id, ctrl)
	a.mu.Unlock()
}

func (a *API) releaseController(id string) {
	a.mu.Lock()
	ctrl := a.controllers[id]
	delete(a.controllers, id)
	a.mu.Unlock()
	if ctrl != nil {
		ctrl.Cancel()
		ctrl.Wait()
	}
}

func (a *API) ensureController(id string) (*runctrl.RunController, error) {
	a.mu.Lock()
	if ctrl := a.controllers[id]; ctrl != nil {
		a.mu.Unlock()
		return ctrl, nil
	}
	a.mu.Unlock()

	rn, err := a.runnerFactory(a.cfg)
	if err != nil {
		return nil, err
	}
	ctrl := a.controllerFactory(a.cfg, a.registry, id, rn)

	a.mu.Lock()
	defer a.mu.Unlock()
	if existing := a.controllers[id]; existing != nil {
		ctrl.Cancel()
		return existing, nil
	}
	a.registerControllerLocked(id, ctrl)
	return ctrl, nil
}

func (a *API) registerControllerLocked(id string, ctrl *runctrl.RunController) {
	ctrl.SetOnTerminal(func() { a.releaseController(id) })
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
}
