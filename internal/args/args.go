package args

import (
	"fmt"
	"strings"
)

type Options struct {
	Prompt       string
	DryRun       bool
	Resume       bool
	Headless     bool
	Verbose      bool
	Help         bool
	Status       bool
	UnknownFlags []string
}

func Parse(args []string) *Options {
	opts := &Options{}
	var promptParts []string

	for _, arg := range args {
		switch arg {
		case "--help", "-h":
			opts.Help = true
		case "--dry-run":
			opts.DryRun = true
		case "--resume":
			opts.Resume = true
		case "--verbose", "-v":
			opts.Verbose = true
		case "run":
			opts.Headless = true
		case "status":
			opts.Status = true
		default:
			if strings.HasPrefix(arg, "-") {
				opts.UnknownFlags = append(opts.UnknownFlags, arg)
			} else {
				promptParts = append(promptParts, arg)
			}
		}
	}

	opts.Prompt = strings.Join(promptParts, " ")
	return opts
}

func (o *Options) Validate() error {
	if o.Help {
		return nil
	}

	if o.Status {
		return nil
	}

	if !o.Resume && o.Prompt == "" {
		return fmt.Errorf("prompt required when not resuming: provide a prompt or use --resume flag")
	}

	if len(o.UnknownFlags) > 0 {
		return fmt.Errorf("unknown flags provided: %v", o.UnknownFlags)
	}

	return nil
}

func HelpText() string {
	return `
Ralph - Autonomous Software Development Agent

Usage:
  ralph "your feature description"               # Full implementation (TUI)
  ralph "your feature description" --dry-run     # Generate PRD only (TUI)
  ralph --resume                                 # Resume from existing prd.json (TUI)
  ralph status                                   # Show current PRD status
  ralph run "your feature description"           # Full implementation (stdout)
  ralph run "your feature description" --dry-run # Generate PRD only (stdout)
  ralph run --resume                             # Resume from existing prd.json (stdout)

Options:
  --dry-run      Generate PRD only, don't implement
  --resume       Resume implementation from existing prd.json
  --verbose, -v  Enable debug logging
  --help, -h     Show this help message

Commands:
  status        Show current PRD status and story progress

Modes:
  (default)    Interactive TUI with progress display
  run          Non-interactive stdout output (for CI/scripts)

AI Models:
  Supports OpenCode and Claude Code CLI models.
  Configure via environment variables:
  - RALPH_MODEL: "opencode/big-pickle" (default), "claude-code/sonnet", "claude-code/haiku", "claude-code/opus"
  - RALPH_MAX_ITERATIONS: Maximum implementation iterations (default: 50)
  - RALPH_RETRY_ATTEMPTS: Max retries per story (default: 3)
  - RALPH_PRD_FILE: PRD filename (default: "prd.json")

Controls (TUI mode):
  q, Ctrl+C    Quit the application

Examples:
  ralph "Add user authentication with login and registration"
  ralph "Create a REST API for managing todos" --dry-run
  ralph --resume
  ralph status
  ralph run "Add unit tests for the API" --dry-run
  ralph run "Add feature" --verbose
`
}
