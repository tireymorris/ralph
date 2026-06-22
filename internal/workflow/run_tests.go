package workflow

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"ralph/internal/prompt"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/logger"
	"ralph/internal/shared/prd"
)

func (e *Executor) effectiveTestCommand(p *prd.PRD) string {
	testCmd := e.cfg.TestCommand
	if p != nil && p.TestCommand != "" {
		testCmd = p.TestCommand
	}
	return strings.TrimSpace(testCmd)
}

func (e *Executor) runTests(p *prd.PRD) (bool, string, error) {
	testCmd := e.effectiveTestCommand(p)
	if testCmd == "" {
		return true, "", nil
	}

	logger.Debug("running test command", "command", testCmd)
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

func (e *Executor) runTestGate(p *prd.PRD) error {
	testCmd := e.effectiveTestCommand(p)
	if testCmd == "" {
		return nil
	}

	success, output, err := e.runTests(p)
	if success {
		if trimmed := strings.TrimSpace(output); trimmed != "" {
			e.emit(EventOutput{Output: Output{Text: "Tests passed:\n" + trimmed}})
		}
		return nil
	}

	testErr := testFailureError(output, err)
	e.emit(EventError{Err: testErr})
	return testErr
}

func testFailureError(output string, err error) error {
	msg := "tests failed"
	if err != nil {
		msg = fmt.Sprintf("tests failed: %v", err)
	}
	if trimmed := strings.TrimSpace(output); trimmed != "" {
		msg += "\n" + trimmed
	}
	return fmt.Errorf("%s", msg)
}

func (e *Executor) runTestGateWithRecovery(ctx context.Context, p *prd.PRD) error {
	if e.effectiveTestCommand(p) == "" {
		return nil
	}

	success, output, err := e.runTests(p)
	if success {
		if trimmed := strings.TrimSpace(output); trimmed != "" {
			e.emit(EventOutput{Output: Output{Text: "Tests passed:\n" + trimmed}})
		}
		return nil
	}
	if !e.cfg.AutoApprove {
		testErr := testFailureError(output, err)
		e.emit(EventError{Err: testErr})
		return testErr
	}

	testErr := testFailureError(output, err)
	for {
		e.emit(EventOutput{Output: Output{Text: "Tests failed; attempting recovery:\n" + testErr.Error(), IsErr: true}})

		recovered, recErr := e.recoverFromReviewFailure(ctx, p, prompt.RecoveryReasonStoryFailure, testErr.Error(), nil)
		if recErr != nil {
			e.emit(EventError{Err: recErr})
			return recErr
		}
		if !recovered {
			e.emit(EventError{Err: testErr})
			return testErr
		}

		success, output, err = e.runTests(p)
		if success {
			if trimmed := strings.TrimSpace(output); trimmed != "" {
				e.emit(EventOutput{Output: Output{Text: "Tests passed:\n" + trimmed}})
			}
			return nil
		}
		testErr = testFailureError(output, err)
		if e.recoveryAttemptsSnapshot() >= constants.MaxRecoveryAttempts {
			e.emit(EventError{Err: testErr})
			return testErr
		}
	}
}
