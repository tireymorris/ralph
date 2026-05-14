package args

import (
	"errors"
	"fmt"
	"strings"
)

type Mode string

const (
	ModeAuto      Mode = "auto"
	ModePRD       Mode = "prd"
	ModeReview    Mode = "review"
	ModeImplement Mode = "implement"
	ModeStatus    Mode = "status"
)

type Options struct {
	Prompt       string
	DryRun       bool
	Resume       bool
	Verbose      bool
	Help         bool
	Mode         Mode
	UnknownFlags []string
}

func Parse(args []string) *Options {
	opts := &Options{Mode: ModeAuto}
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
			opts.Mode = ModeAuto
		case "status":
			opts.Mode = ModeStatus
		case "prd":
			opts.Mode = ModePRD
		case "review":
			opts.Mode = ModeReview
		case "implement":
			opts.Mode = ModeImplement
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
	if o.Mode == ModeStatus || o.Mode == ModeReview || o.Mode == ModeImplement {
		return nil
	}
	if !o.Resume && o.Prompt == "" {
		return errors.New("prompt required when not resuming: provide a prompt or use --resume flag")
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
  ralph run "your feature description"           # TUI mode
  ralph run --resume                             # Resume (TUI)
  ralph prd "your feature description"           # Generate PRD only, no implementation
  ralph prd "your feature description" --verbose # Generate PRD with debug logging
  ralph review                                   # Review existing PRD
  ralph review --verbose                         # Review with debug logging
  ralph implement                                # Implement stories from existing PRD
  ralph implement --verbose                      # Implement with debug logging

Options:
  --dry-run      Generate PRD only, don't implement
  --resume       Resume implementation from existing prd.json
  --verbose, -v  Enable debug logging
  --help, -h     Show this help message
`
}
