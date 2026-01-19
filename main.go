package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/args"
	"ralph/internal/cli"
	"ralph/internal/config"
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

	cfg := config.Load()

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
		return cli.NewRunner(cfg, opts.Prompt, opts.DryRun, opts.Resume).Run()
	}

	return runTUI(cfg, opts)
}

func runTUI(cfg *config.Config, opts *args.Options) int {
	model := tui.NewModel(cfg, opts.Prompt, opts.DryRun, opts.Resume)
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
