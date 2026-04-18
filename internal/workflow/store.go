package workflow

import (
	"ralph/internal/config"
	"ralph/internal/prd"
)

// PRDStore abstracts PRD persistence so workflow can be tested without disk I/O.
type PRDStore interface {
	Load(cfg *config.Config) (*prd.PRD, error)
	Save(cfg *config.Config, p *prd.PRD) error
	Exists(cfg *config.Config) bool
}

type defaultPRDStore struct{}

func (defaultPRDStore) Load(cfg *config.Config) (*prd.PRD, error) {
	return prd.Load(cfg)
}

func (defaultPRDStore) Save(cfg *config.Config, p *prd.PRD) error {
	return prd.Save(cfg, p)
}

func (defaultPRDStore) Exists(cfg *config.Config) bool {
	return prd.Exists(cfg)
}
