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

	// Create temp dir and invalid config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "ralph.config.json")
	os.WriteFile(configPath, []byte("invalid json"), 0644)

	// Get absolute path to binary
	binaryPath, _ := filepath.Abs("ralph-test")

	// Run with invalid config
	cmd = exec.Command(binaryPath, "test prompt")
	cmd.Dir = tmpDir
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
