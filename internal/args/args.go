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

	if !o.Resume && o.Prompt == "" {
		return fmt.Errorf("please provide a prompt or use --resume")
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
  ralph run "your feature description"           # Full implementation (stdout)
  ralph run "your feature description" --dry-run # Generate PRD only (stdout)
  ralph run --resume                             # Resume from existing prd.json (stdout)

Options:
  --dry-run      Generate PRD only, don't implement
  --resume       Resume implementation from existing prd.json
  --verbose, -v  Enable debug logging
  --help, -h     Show this help message

Modes:
  (default)    Interactive TUI with progress display
  run          Non-interactive stdout output (for CI/scripts)

AI Models:
  Supports OpenCode and Claude Code CLI models.
  Configure model in ralph.config.json:
  - OpenCode: "opencode/big-pickle" (default), "opencode/glm-4.7-free", etc.
  - Claude Code: "claude-code/claude-3.5-sonnet", "claude-code/claude-3.5-haiku", etc.

Controls (TUI mode):
  q, Ctrl+C    Quit the application

Examples:
  ralph "Add user authentication with login and registration"
  ralph "Create a REST API for managing todos" --dry-run
  ralph --resume
  ralph run "Add unit tests for the API" --dry-run
  ralph run "Add feature" --verbose
`
}
