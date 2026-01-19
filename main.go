package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/args"
	"ralph/internal/cli"
	"ralph/internal/config"
	"ralph/internal/logger"
	"ralph/internal/prd"
	"ralph/internal/tui"
)

func main() {
	os.Exit(run())
}

func run() int {
	opts := args.Parse(os.Args[1:])

	if opts.Help {
		fmt.Print(args.HelpText())
		return 0
	}

	if err := opts.Validate(); err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Print(args.HelpText())
		return 1
	}

	// Warn about unknown flags
	for _, flag := range opts.UnknownFlags {
		fmt.Fprintf(os.Stderr, "Warning: unknown flag %q (ignored)\n", flag)
	}

	// Initialize logger with verbose flag
	logger.Init(opts.Verbose)
	logger.Debug("starting ralph", "verbose", opts.Verbose)

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Print(args.HelpText())
		return 1
	}
	logger.Debug("config loaded", "model", cfg.Model, "max_iterations", cfg.MaxIterations)

	if opts.Resume {
		if !prd.Exists(cfg) {
			fmt.Printf("Error: No %s found to resume from\n", cfg.PRDFile)
			fmt.Println("Run ralph with a prompt first to generate a PRD")
			return 1
		}
		if _, err := prd.Load(cfg); err != nil {
			fmt.Printf("Error: %v\n", err)
			return 1
		}
	}

	if opts.Headless {
		return cli.NewRunner(cfg, opts.Prompt, opts.DryRun, opts.Resume, opts.Verbose).Run()
	}

	return runTUI(cfg, opts)
}

func runTUI(cfg *config.Config, opts *args.Options) int {
	model := tui.NewModel(cfg, opts.Prompt, opts.DryRun, opts.Resume, opts.Verbose)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return 1
	}

	if m, ok := finalModel.(*tui.Model); ok {
		return m.ExitCode()
	}

	return 0
}
