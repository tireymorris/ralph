package workflow

import (
	"fmt"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/workdir"
)

var currentBranchName = workdir.CurrentBranchName
var isDefaultBranch = workdir.IsDefaultBranch
var checkoutBranch = workdir.CheckoutBranch
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
	if !isDefaultBranch(branchName, d.cfg.DefaultBranches) {
		if p.BranchName != branchName {
			p.BranchName = branchName
			if err := savePRD(d.cfg, p); err != nil {
				return fmt.Errorf("save PRD branch %q: %w", branchName, err)
			}
			d.mu.Lock()
			d.currentPRD = p
			d.mu.Unlock()
		}
		return nil
	}
	if err := checkoutBranch(d.cfg.WorkDir, p.BranchName); err != nil {
		return fmt.Errorf("checkout PRD branch %q: %w", p.BranchName, err)
	}
	return nil
}
