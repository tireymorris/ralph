package workflow

import (
	"fmt"

	"github.com/tireymorris/ralph/internal/shared/config"
	"github.com/tireymorris/ralph/internal/shared/prd"
	"github.com/tireymorris/ralph/internal/shared/workdir"
)

var currentBranchName = workdir.CurrentBranchName
var savePRD = func(cfg *config.Config, p *prd.PRD) error {
	return prd.Save(cfg, p)
}

func (d *Driver) prepareImplementationBranch(p *prd.PRD) error {
	if err := workdir.ValidateGit(d.cfg.WorkDir); err != nil {
		return nil
	}
	branchName, err := currentBranchName(d.cfg.WorkDir)
	if err != nil {
		return fmt.Errorf("detect active branch: %w", err)
	}
	if p.BranchName == branchName {
		return nil
	}
	p.BranchName = branchName
	if err := savePRD(d.cfg, p); err != nil {
		return fmt.Errorf("save PRD branch %q: %w", branchName, err)
	}
	d.mu.Lock()
	d.currentPRD = p
	d.mu.Unlock()
	return nil
}
