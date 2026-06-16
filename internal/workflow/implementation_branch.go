package workflow

import (
	"fmt"
	"os/exec"
	"strings"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/workdir"
)

var currentBranchName = workdir.CurrentBranchName
var isDefaultBranch = workdir.IsDefaultBranch
var checkoutBranch = runGitCheckout
var savePRD = func(cfg *config.Config, p *prd.PRD) error {
	return prd.Save(cfg, p)
}

func runGitCheckout(workDir, branchName string) error {
	cmd := exec.Command("git", "checkout", "-B", branchName)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout %s in %s: %w: %s", branchName, workDir, err, strings.TrimSpace(string(out)))
	}
	return nil
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
