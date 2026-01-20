//go:build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
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

	// Create temp dir with valid config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "ralph.config.json")
	configContent := `{"max_iterations":5,"retry_attempts":3}`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Get absolute path to binary
	binaryPath, _ := filepath.Abs("ralph-test")

	// Run 'ralph run "test" --dry-run'
	cmd = exec.Command(binaryPath, "run", "test", "--dry-run")
	cmd.Dir = tmpDir
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

	// Create temp dir with valid config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "ralph.config.json")
	configContent := `{"max_iterations":5,"retry_attempts":3}`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Get absolute path to binary
	binaryPath, _ := filepath.Abs("ralph-test")

	// Run 'ralph "test prompt" --dry-run' with a simulated TUI interaction
	// We'll use expect or a similar tool to interact with the TUI
	cmd = exec.Command("expect", "-c", `
		spawn "`+binaryPath+`" "test prompt" --dry-run
		expect {
			"Phase 1: PRD Generation" { send "q\r" }
			timeout { exit 1 }
		}
		expect eof
	`)
	cmd.Dir = tmpDir
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

	// Create temp dir with valid config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "ralph.config.json")
	configContent := `{"max_iterations":5,"retry_attempts":3}`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Get absolute path to binary
	binaryPath, _ := filepath.Abs("ralph-test")

	// Run with invalid prompt that causes opencode failure
	cmd = exec.Command(binaryPath, "run", "invalid prompt that should cause parsing failure")
	cmd.Dir = tmpDir
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

func TestIntegrationDocumentation(t *testing.T) {
	// 1) Run 'go doc ./...' and assert no missing documentation warnings
	docCmd := exec.Command("go", "doc", "./...")
	docOutput, _ := docCmd.CombinedOutput()
	docOutputStr := string(docOutput)

	// Check for documentation warnings (go doc doesn't fail on missing docs, but we can check output)
	if strings.Contains(docOutputStr, "undocumented") || strings.Contains(docOutputStr, "missing") {
		t.Errorf("Found missing documentation warnings in 'go doc ./...' output: %s", docOutputStr)
	}

	// 2) Run 'go build' and assert compilation succeeds
	buildCmd := exec.Command("go", "build", "-o", "ralph-test", ".")
	buildOutput, buildErr := buildCmd.CombinedOutput()
	if buildErr != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", buildErr, buildOutput)
	}
	defer os.Remove("ralph-test")

	// 3) Get absolute path to binary
	binaryPath, _ := filepath.Abs("ralph-test")

	// 4) Run 'ralph --help' and assert help text is comprehensive
	helpCmd := exec.Command(binaryPath, "--help")
	helpOutput, helpErr := helpCmd.CombinedOutput()
	if helpErr != nil && helpCmd.ProcessState == nil {
		t.Fatalf("Command failed: %v", helpErr)
	}
	exitCode := helpCmd.ProcessState.ExitCode()

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for --help, got %d", exitCode)
	}

	helpOutputStr := string(helpOutput)

	// 5) Assert help text is comprehensive
	requiredSections := []string{
		"Usage:",
		"Options:",
		"--dry-run",
		"--resume",
		"--verbose",
		"--help",
		"Examples:",
	}

	for _, section := range requiredSections {
		if !strings.Contains(helpOutputStr, section) {
			t.Errorf("Help text missing required section '%s'. Full output: %s", section, helpOutputStr)
		}
	}

	// Assert help contains multiple examples
	if strings.Count(helpOutputStr, "ralph ") < 3 {
		t.Errorf("Help text should contain multiple usage examples, got: %s", helpOutputStr)
	}
}

func TestIntegrationGitHubActionsWorkflow(t *testing.T) {
	// 1) Verify .github/workflows/ directory exists
	if _, err := os.Stat(".github/workflows"); os.IsNotExist(err) {
		t.Errorf("Expected .github/workflows directory to exist")
	}

	// 2) Assert that a workflow file (ci.yml) is present
	workflowPath := ".github/workflows/ci.yml"
	if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
		t.Errorf("Expected workflow file %s to exist", workflowPath)
	}

	// Read the workflow file
	workflowContent, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("Failed to read workflow file: %v", err)
	}

	// Check YAML syntax by unmarshaling
	var workflow map[interface{}]interface{}
	if err := yaml.Unmarshal(workflowContent, &workflow); err != nil {
		t.Errorf("Workflow file is not valid YAML: %v", err)
	}

	// Assert required keys
	workflowStr := string(workflowContent)
	requiredKeys := []string{
		"name:",
		"on:",
		"push:",
		"branches: [ main ]",
		"pull_request:",
		"jobs:",
		"test:",
		"runs-on: ubuntu-latest",
		"steps:",
		"actions/checkout@v4",
		"actions/setup-go@v4",
		"go-version: 1.24.0",
		"run: go test ./... -cover",
	}

	for _, key := range requiredKeys {
		if !strings.Contains(workflowStr, key) {
			t.Errorf("Workflow file missing required key: %s", key)
		}
	}

	// 3) Run 'go test ./... -cover' locally to ensure tests pass and coverage.out is generated
	cmd := exec.Command("go", "test", "./...", "-cover")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go test failed: %v\nOutput: %s", err, output)
	}

	// Assert coverage output is generated (look for "coverage:" in output)
	if !strings.Contains(string(output), "coverage:") {
		t.Errorf("Expected coverage output, but not found in: %s", output)
	}

	// Assert all tests passed (no "FAIL" in output)
	if strings.Contains(string(output), "FAIL") {
		t.Errorf("Some tests failed: %s", output)
	}
}

func TestIntegrationCommentRemoval(t *testing.T) {
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

	// Run --help to verify runtime functionality is unchanged
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

	// Verify no runtime panics or errors after comment removal
	if strings.Contains(outputStr, "panic") || strings.Contains(outputStr, "runtime error") {
		t.Errorf("Runtime panic detected in output: %s", outputStr)
	}
}
