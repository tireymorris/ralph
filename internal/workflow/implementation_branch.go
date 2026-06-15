package workflow

import (
	"fmt"
	"os/exec"
	"strings"

	"ralph/internal/shared/prd"
	"ralph/internal/shared/workdir"
)

var currentBranchName = workdir.CurrentBranchName
var isFeatureBranch = workdir.IsFeatureBranch
var checkoutBranch = runGitCheckout

func runGitCheckout(workDir, branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
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
	if isFeatureBranch(branchName) {
		return nil
	}
	if err := checkoutBranch(d.cfg.WorkDir, p.BranchName); err != nil {
		return fmt.Errorf("checkout PRD branch %q: %w", p.BranchName, err)
	}
	return nil
}
