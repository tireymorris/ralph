//go:build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegrationHelp(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "ralph-test", ".")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}
	defer os.Remove("ralph-test")

	// Get absolute path to binary
	binaryPath, _ := filepath.Abs("ralph-test")

	// Run --help
	cmd = exec.Command(binaryPath, "--help")
	output, err = cmd.CombinedOutput()
	if err != nil && cmd.ProcessState == nil {
		t.Fatalf("Command failed: %v", err)
	}
	exitCode := cmd.ProcessState.ExitCode()

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Usage") {
		t.Errorf("Expected output to contain 'Usage', got: %s", outputStr)
	}
}

func TestIntegrationInvalidConfig(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "ralph-test", ".")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}
	defer os.Remove("ralph-test")

	// Get absolute path to binary
	binaryPath, _ := filepath.Abs("ralph-test")

	// Run with invalid config (invalid model)
	cmd = exec.Command(binaryPath, "test prompt")
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "RALPH_MODEL=invalid-model")
	output, err = cmd.CombinedOutput()
	if err != nil && cmd.ProcessState == nil {
		t.Fatalf("Command failed: %v", err)
	}
	exitCode := cmd.ProcessState.ExitCode()

	if exitCode == 0 {
		t.Errorf("Expected non-zero exit code for invalid config, got 0")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Error:") {
		t.Errorf("Expected output to contain 'Error:', got: %s", outputStr)
	}
}

func TestIntegrationDryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Build the binary
	cmd := exec.Command("go", "build", "-o", "ralph-test", ".")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}
	defer os.Remove("ralph-test")

	// Get absolute path to binary
	binaryPath, _ := filepath.Abs("ralph-test")

	// Run 'ralph run "test" --dry-run' with env vars
	tmpDir := t.TempDir()
	cmd = exec.Command(binaryPath, "run", "test", "--dry-run")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "RALPH_MAX_ITERATIONS=5", "RALPH_RETRY_ATTEMPTS=3")
	output, err = cmd.CombinedOutput()
	if err != nil && cmd.ProcessState == nil {
		t.Fatalf("Command failed: %v", err)
	}
	exitCode := cmd.ProcessState.ExitCode()
	outputStr := string(output)

	// Assert no runtime panics (panic would show in output or exit code)
	if strings.Contains(outputStr, "panic") || strings.Contains(outputStr, "runtime error") {
		t.Errorf("Runtime panic detected in output: %s", outputStr)
	}

	// Assert PRD generation completes without errors
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for successful dry run, got %d. Output: %s", exitCode, outputStr)
	}

	// Assert PRD generation completed
	if !strings.Contains(outputStr, "stories") || !strings.Contains(outputStr, "Dry run complete") {
		t.Errorf("Expected PRD generation success messages, got: %s", outputStr)
	}
}

func TestIntegrationTUIDryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Build the binary
	cmd := exec.Command("go", "build", "-o", "ralph-test", ".")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}
	defer os.Remove("ralph-test")

	// Get absolute path to binary
	binaryPath, _ := filepath.Abs("ralph-test")

	// Run 'ralph "test prompt" --dry-run' with a simulated TUI interaction
	// We'll use expect or a similar tool to interact with the TUI
	tmpDir := t.TempDir()
	cmd = exec.Command("expect", "-c", `
		spawn "`+binaryPath+`" "test prompt" --dry-run
		expect {
			"Phase 1: PRD Generation" { send "q\r" }
			timeout { exit 1 }
		}
		expect eof
	`)
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "RALPH_MAX_ITERATIONS=5", "RALPH_RETRY_ATTEMPTS=3")
	output, err = cmd.CombinedOutput()
	outputStr := string(output)

	// Check for errors
	if err != nil {
		t.Errorf("TUI interaction failed: %v\nOutput: %s", err, outputStr)
	}

	// Assert TUI displays PRD generation phase correctly
	if !strings.Contains(outputStr, "Phase 1: PRD Generation") {
		t.Errorf("Expected TUI to display 'Phase 1: PRD Generation', got: %s", outputStr)
	}

	// Assert clean exit with no errors (expect script should exit cleanly)
	if strings.Contains(outputStr, "Error:") || strings.Contains(outputStr, "panic") {
		t.Errorf("Expected clean exit with no errors, but found errors in output: %s", outputStr)
	}
}

func TestIntegrationOpencodeFailure(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "ralph-test", ".")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}
	defer os.Remove("ralph-test")

	// Get absolute path to binary
	binaryPath, _ := filepath.Abs("ralph-test")

	// Run with invalid prompt that causes opencode failure
	tmpDir := t.TempDir()
	cmd = exec.Command(binaryPath, "run", "invalid prompt that should cause parsing failure")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "RALPH_MAX_ITERATIONS=5", "RALPH_RETRY_ATTEMPTS=3")
	output, err = cmd.CombinedOutput()
	if err != nil && cmd.ProcessState == nil {
		t.Fatalf("Command failed: %v", err)
	}
	exitCode := cmd.ProcessState.ExitCode()

	// Assert non-zero exit code for failure
	if exitCode == 0 {
		t.Errorf("Expected non-zero exit code for opencode failure, got 0")
	}

	outputStr := string(output)

	// Assert appropriate error message displayed
	if !strings.Contains(outputStr, "Error:") {
		t.Errorf("Expected output to contain 'Error:', got: %s", outputStr)
	}

	// Check for structured error messages
	if !strings.Contains(outputStr, "opencode") && !strings.Contains(outputStr, "PRD") && !strings.Contains(outputStr, "git") {
		t.Errorf("Expected output to contain structured error message, got: %s", outputStr)
	}

	// Run with --verbose to verify detailed error logging
	cmd = exec.Command(binaryPath, "run", "invalid prompt that should cause parsing failure", "--verbose")
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "RALPH_MAX_ITERATIONS=5", "RALPH_RETRY_ATTEMPTS=3")
	output, err = cmd.CombinedOutput()
	if err != nil && cmd.ProcessState == nil {
		t.Fatalf("Command failed: %v", err)
	}
	outputStrVerbose := string(output)

	// Assert detailed error logging in verbose mode
	if !strings.Contains(outputStrVerbose, "error") {
		t.Errorf("Expected verbose output to contain error details, got: %s", outputStrVerbose)
	}
}
