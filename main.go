package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/config"
	"ralph/internal/prd"
	"ralph/internal/tui"
)

const (
	exitSuccess = 0
	exitFailure = 1
)

func main() {
	os.Exit(run())
}

func run() int {
	args := os.Args[1:]

	// Parse flags
	dryRun := false
	resume := false
	runMode := false
	var promptParts []string

	for _, arg := range args {
		switch arg {
		case "--help", "-h":
			showHelp()
			return exitSuccess
		case "--dry-run":
			dryRun = true
		case "--resume":
			resume = true
		case "run":
			runMode = true
		default:
			if !strings.HasPrefix(arg, "-") {
				promptParts = append(promptParts, arg)
			}
		}
	}

	prompt := strings.Join(promptParts, " ")

	// Validate arguments
	if !resume && prompt == "" && !runMode {
		fmt.Println("Error: Please provide a prompt or use --resume")
		showHelp()
		return exitFailure
	}

	cfg := config.Load()

	// If run mode is specified, use stdout output
	if runMode {
		fmt.Printf("Running in stdout mode with model: %s\n", cfg.Model)
		fmt.Printf("Prompt: %s\n", prompt)
		fmt.Println("Configuration loaded successfully")
		// In a real implementation, you'd process the prompt and output results here
		// For now we're just showing that the configuration loaded
		return exitSuccess
	}

	// Check for resume without PRD file
	if resume && !prd.Exists(cfg) {
		fmt.Printf("Error: No %s found to resume from\n", cfg.PRDFile)
		fmt.Println("Run ralph with a prompt first to generate a PRD")
		return exitFailure
	}

	// Validate PRD if resuming
	if resume {
		_, err := prd.Load(cfg)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return exitFailure
		}
	}

	// Create and run the TUI
	model := tui.NewModel(cfg, prompt, dryRun, resume)

	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return exitFailure
	}

	// Determine exit code based on final state
	if m, ok := finalModel.(*tui.Model); ok {
		return m.ExitCode()
	}

	return exitSuccess
}

func showHelp() {
	help := `
Ralph - Autonomous Software Development Agent (TUI)

Usage:
  ralph "your feature description"           # Full implementation
  ralph "your feature description" --dry-run # Generate PRD only
  ralph --resume                             # Resume from existing prd.json
  ralph run "your feature description"       # Run with stdout output

Options:
  --dry-run    Generate PRD only, don't implement
  --resume     Resume implementation from existing prd.json
  --help, -h   Show this help message

Controls:
  q, Ctrl+C    Quit the application

Examples:
  ralph "Add user authentication with login and registration"
  ralph "Create a REST API for managing todos" --dry-run
  ralph --resume
  ralph run "Test stdout mode"
`
	fmt.Println(help)
}
