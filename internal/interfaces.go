package internal

import (
	"context"

	"ralph/internal/prd"
)

type WorkflowRunner interface {
	RunGenerate(ctx context.Context, prompt string) (*prd.PRD, error)
	RunLoad(ctx context.Context) (*prd.PRD, error)
	RunImplementation(ctx context.Context, p *prd.PRD) error
}
