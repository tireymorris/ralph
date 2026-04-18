package workflow

import (
	"context"
	"fmt"

	"ralph/internal/logger"
	"ralph/internal/prd"
)

func (e *Executor) RunLoad(ctx context.Context) (*prd.PRD, error) {
	_ = ctx
	p, err := e.store.Load(e.cfg)
	if err != nil {
		e.emit(EventError{Err: fmt.Errorf("failed to load PRD %s: %w", e.cfg.PRDFile, err)})
		return nil, fmt.Errorf("failed to load PRD %s: %w", e.cfg.PRDFile, err)
	}

	logger.Debug("PRD loaded", "project", p.ProjectName, "stories", len(p.Stories))
	e.emit(EventPRDLoaded{PRD: p})
	e.emit(EventPRDReview{PRD: p})
	return p, nil
}
