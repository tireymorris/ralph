package workflow

import (
	"os/exec"

	"ralph/internal/logger"
	"ralph/internal/prd"
)

func (e *Executor) runTests(p *prd.PRD) (bool, string, error) {
	testCmd := e.cfg.TestCommand
	if p != nil && p.TestCommand != "" {
		testCmd = p.TestCommand
		logger.Debug("using PRD test_command", "command", testCmd)
	}
	cmd := exec.Command("sh", "-c", testCmd)
	if e.cfg.WorkDir != "" {
		cmd.Dir = e.cfg.WorkDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, string(output), err
	}
	return true, string(output), nil
}
