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
	return `Ralph - Autonomous Software Development Agent

Usage:
  ralph "your feature description"               # TUI mode
  ralph "your feature description" --dry-run     # Generate PRD only
  ralph --resume                                 # Resume from existing prd.json
  ralph status                                   # Show current PRD status
  ralph run "your feature description"           # Headless/stdout mode
  ralph run --resume                             # Resume (headless)

Options:
  --dry-run      Generate PRD only, don't implement
  --resume       Resume implementation from existing prd.json
  --verbose, -v  Enable debug logging
  --help, -h     Show this help message
`
}
