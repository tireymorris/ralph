package workflow

import (
	"context"

	"ralph/internal/shared/prd"
)

func (e *Executor) RunCleanup(ctx context.Context, p *prd.PRD) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}
