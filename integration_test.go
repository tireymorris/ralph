//go:build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func requireBinary(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("skipping integration test: %s not found in PATH", name)
	}
}

func requireTTY(t *testing.T) {
	t.Helper()
	f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		t.Skipf("skipping integration test: no TTY available: %v", err)
	}
	_ = f.Close()
}

func buildTestBinary(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("go", "build", "-o", "ralph-test", ".")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build binary: %v\nOutput: %s", err, output)
	}
	t.Cleanup(func() { _ = os.Remove("ralph-test") })
	binaryPath, err := filepath.Abs("ralph-test")
	if err != nil {
		t.Fatalf("abs binary path: %v", err)
	}
	return binaryPath
}

func TestIntegrationHelp(t *testing.T) {
	binaryPath := buildTestBinary(t)

	cmd := exec.Command(binaryPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil && cmd.ProcessState == nil {
		t.Fatalf("command failed: %v", err)
	}
	if got := cmd.ProcessState.ExitCode(); got != 0 {
		t.Errorf("exit code = %d, want 0", got)
	}
	if !strings.Contains(string(output), "Usage") {
		t.Errorf("expected output to contain Usage, got: %s", output)
	}
}

func TestIntegrationInvalidConfig(t *testing.T) {
	binaryPath := buildTestBinary(t)

	cmd := exec.Command(binaryPath, "test prompt")
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "RALPH_RUNNER=invalid-runner")
	output, err := cmd.CombinedOutput()
	if err != nil && cmd.ProcessState == nil {
		t.Fatalf("command failed: %v", err)
	}
	if got := cmd.ProcessState.ExitCode(); got == 0 {
		t.Errorf("exit code = %d, want non-zero", got)
	}
	if !strings.Contains(string(output), "Error") {
		t.Errorf("expected output to contain Error, got: %s", output)
	}
}

func TestIntegrationDryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	requireBinary(t, "opencode")
	requireTTY(t)
	binaryPath := buildTestBinary(t)

	cmd := exec.Command(binaryPath, "run", "test", "--dry-run")
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "RALPH_RUNNER=opencode")
	output, err := cmd.CombinedOutput()
	if err != nil && cmd.ProcessState == nil {
		t.Fatalf("command failed: %v", err)
	}
	outputStr := string(output)

	if strings.Contains(outputStr, "panic") || strings.Contains(outputStr, "runtime error") {
		t.Errorf("runtime panic detected in output: %s", outputStr)
	}
	if got := cmd.ProcessState.ExitCode(); got != 0 {
		t.Errorf("exit code = %d, want 0. Output: %s", got, outputStr)
	}
	if !strings.Contains(outputStr, "stories") || !strings.Contains(outputStr, "Dry run complete") {
		t.Errorf("expected PRD generation success messages, got: %s", outputStr)
	}
}

func TestIntegrationTUIDryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	requireBinary(t, "expect")
	requireBinary(t, "opencode")
	requireTTY(t)
	binaryPath := buildTestBinary(t)

	cmd := exec.Command("expect", "-c", `
		spawn "`+binaryPath+`" "test prompt" --dry-run
		expect {
			"Phase 1: PRD Generation" { send "q\r" }
			timeout { exit 1 }
		}
		expect eof
	`)
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "RALPH_RUNNER=opencode")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		t.Errorf("TUI interaction failed: %v\nOutput: %s", err, outputStr)
	}
	if !strings.Contains(outputStr, "Phase 1: PRD Generation") {
		t.Errorf("expected TUI phase, got: %s", outputStr)
	}
	if strings.Contains(outputStr, "Error:") || strings.Contains(outputStr, "panic") {
		t.Errorf("expected clean exit with no errors, got: %s", outputStr)
	}
}

func TestIntegrationOpencodeFailure(t *testing.T) {
	requireBinary(t, "opencode")
	requireTTY(t)
	binaryPath := buildTestBinary(t)

	cmd := exec.Command(binaryPath, "run", "invalid prompt that should cause parsing failure")
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "RALPH_RUNNER=opencode")
	output, err := cmd.CombinedOutput()
	if err != nil && cmd.ProcessState == nil {
		t.Fatalf("command failed: %v", err)
	}
	if got := cmd.ProcessState.ExitCode(); got == 0 {
		t.Errorf("exit code = %d, want non-zero", got)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Error:") {
		t.Errorf("expected output to contain Error:, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "opencode") && !strings.Contains(outputStr, "PRD") && !strings.Contains(outputStr, "git") {
		t.Errorf("expected structured error message, got: %s", outputStr)
	}

	cmd = exec.Command(binaryPath, "run", "invalid prompt that should cause parsing failure", "--verbose")
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "RALPH_RUNNER=opencode")
	output, err = cmd.CombinedOutput()
	if err != nil && cmd.ProcessState == nil {
		t.Fatalf("command failed: %v", err)
	}
	if !strings.Contains(string(output), "error") {
		t.Errorf("expected verbose output to contain error details, got: %s", output)
	}
}
