package cli

import (
	"fmt"

	"ralph/internal/args"
	"ralph/internal/config"
	"ralph/internal/prd"
)

// RunNonTUI handles status and headless modes. If it returns handled=false, the caller should start the TUI.
func RunNonTUI(cfg *config.Config, opts *args.Options) (exitCode int, handled bool) {
	if opts.Status {
		if err := RunStatus(cfg); err != nil {
			return 1, true
		}
		return 0, true
	}
	if opts.Headless {
		cmd := HeadlessCommand{
			Cfg:     cfg,
			Prompt:  opts.Prompt,
			DryRun:  opts.DryRun,
			Resume:  opts.Resume,
			Verbose: opts.Verbose,
		}
		return cmd.Run(), true
	}
	return 0, false
}

// ValidateResume checks that a PRD exists and loads when --resume is set.
func ValidateResume(cfg *config.Config, resume bool) error {
	if !resume {
		return nil
	}
	if !prd.Exists(cfg) {
		return fmt.Errorf("no %s found to resume from (run ralph with a prompt first to generate a PRD)", cfg.PRDFile)
	}
	if _, err := prd.Load(cfg); err != nil {
		return fmt.Errorf("loading existing PRD %s: %w", cfg.PRDFile, err)
	}
	return nil
}
