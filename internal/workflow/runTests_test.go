package workflow

import (
	"os/exec"
	"strings"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/logger"
	"ralph/internal/shared/prd"
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

func TestRunTestsReturnsTrueOnSuccess(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.TestCommand = "echo success"
	cfg.WorkDir = ""

	e := NewExecutor(cfg, nil)

	success, output, err := e.runTests(nil)

	if err != nil {
		t.Fatalf("runTests() error = %v, want nil", err)
	}
	if !success {
		t.Error("runTests() success = false, want true")
	}
	if output == "" {
		t.Error("runTests() output = empty, want output")
	}
}

func TestRunTestsReturnsFalseOnFailure(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.TestCommand = "echo error msg && exit 1"
	cfg.WorkDir = ""

	e := NewExecutor(cfg, nil)

	success, output, err := e.runTests(nil)

	if err == nil {
		t.Error("runTests() error = nil, want error when command fails")
	}
	if success {
		t.Error("runTests() success = true, want false")
	}
	if !strings.Contains(output, "error msg") {
		t.Errorf("runTests() output = %q, want containing %q", output, "error msg")
	}
}

func TestRunTestsUsesConfigTestCommand(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.TestCommand = "go version"
	cfg.WorkDir = ""

	e := NewExecutor(cfg, nil)

	success, output, err := e.runTests(nil)

	if err != nil {
		t.Fatalf("runTests() error = %v, want nil", err)
	}
	if !success {
		t.Error("runTests() success = false, want true for go version")
	}
	if !strings.Contains(output, "go") {
		t.Errorf("runTests() output = %q, want containing %q", output, "go")
	}
}

func TestRunTestsExecutesInWorkDir(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.TestCommand = "pwd"
	cfg.WorkDir = "/tmp"

	e := NewExecutor(cfg, nil)

	success, output, err := e.runTests(nil)

	if err != nil {
		t.Fatalf("runTests() error = %v, want nil", err)
	}
	if !success {
		t.Error("runTests() success = false, want true")
	}
	if output == "" {
		t.Error("runTests() output = empty, want pwd output")
	}
}

func TestRunTestsUsesPRDTestCommand(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.TestCommand = "echo config default"
	cfg.WorkDir = ""

	e := NewExecutor(cfg, nil)

	p := &prd.PRD{
		TestCommand: "echo prd override",
	}

	success, output, err := e.runTests(p)

	if err != nil {
		t.Fatalf("runTests() error = %v, want nil", err)
	}
	if !success {
		t.Error("runTests() success = false, want true")
	}
	if !strings.Contains(output, "prd override") {
		t.Errorf("runTests() output = %q, want containing %q", output, "prd override")
	}
}
