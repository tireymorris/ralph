package app

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/args"
	"ralph/internal/shared/config"
	"ralph/internal/shared/logger"
	sharedprd "ralph/internal/shared/prd"
	"ralph/internal/status"
	"ralph/internal/tui"
)

func Run(argv []string) int {
	opts := args.Parse(argv)
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
	logger.Debug("config loaded", "model", cfg.Model)

	if err := ValidateResume(cfg, opts.Resume); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if opts.Status {
		return RunStatus(cfg)
	}
	return RunTUI(cfg, opts.Prompt, opts.DryRun, opts.Resume, opts.Verbose)
}

func RunTUI(cfg *config.Config, prompt string, dryRun, resume, verbose bool) int {
	model := tui.NewModel(cfg, prompt, dryRun, resume, verbose)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
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

func RunStatus(cfg *config.Config) int {
	if err := status.Display(cfg); err != nil {
		return 1
	}
	return 0
}

func ValidateResume(cfg *config.Config, resume bool) error {
	if !resume {
		return nil
	}
	exists, err := sharedprd.Exists(cfg)
	if err != nil {
		return fmt.Errorf("checking for existing PRD %s: %w", cfg.PRDFile, err)
	}
	if !exists {
		return fmt.Errorf("no %s found to resume from (run ralph with a prompt first to generate a PRD)", cfg.PRDFile)
	}
	if _, err := sharedprd.Load(cfg); err != nil {
		return fmt.Errorf("loading existing PRD %s: %w", cfg.PRDFile, err)
	}
	return nil
}
