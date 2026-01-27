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
	"ralph/internal/status"
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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Print(args.HelpText())
		return 1
	}

	for _, flag := range opts.UnknownFlags {
		fmt.Fprintf(os.Stderr, "Warning: unknown flag %q (ignored)\n", flag)
	}

	logger.Init(opts.Verbose)
	logger.Debug("starting ralph", "verbose", opts.Verbose)

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		fmt.Print(args.HelpText())
		return 1
	}
	logger.Debug("config loaded", "model", cfg.Model, "max_iterations", cfg.MaxIterations)

	if opts.Status {
		if err := status.Display(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error displaying status: %v\n", err)
			return 1
		}
		return 0
	}

	if opts.Resume {
		if !prd.Exists(cfg) {
			fmt.Fprintf(os.Stderr, "Error: No %s found to resume from\n", cfg.PRDFile)
			fmt.Println("Run ralph with a prompt first to generate a PRD")
			return 1
		}
		if _, err := prd.Load(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading existing PRD %s: %v\n", cfg.PRDFile, err)
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
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		return 1
	}

	if m, ok := finalModel.(*tui.Model); ok {
		return m.ExitCode()
	}

	return 0
}
