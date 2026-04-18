package workflow

import (
	"strings"
	"testing"

	"ralph/internal/config"
)

func TestRunTestsReturnsTrueOnSuccess(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.TestCommand = "echo success"
	cfg.WorkDir = ""

	e := NewExecutor(cfg, nil)

	success, output, err := e.runTests()

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

	success, output, err := e.runTests()

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

	success, output, err := e.runTests()

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

	success, output, err := e.runTests()

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