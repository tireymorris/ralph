package args

import (
	"fmt"
	"strconv"
	"strings"
)

type Options struct {
	Prompt       string
	DryRun       bool
	Resume       bool
	Verbose      bool
	Help         bool
	Status       bool
	Web          bool
	WebPort      int
	UnknownFlags []string
}

const defaultWebPort = 8080

func Parse(args []string) *Options {
	opts := &Options{}
	var promptParts []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--help", "-h":
			opts.Help = true
		case "--dry-run":
			opts.DryRun = true
		case "--resume":
			opts.Resume = true
		case "--verbose", "-v":
			opts.Verbose = true
		case "status":
			opts.Status = true
		case "web":
			opts.Web = true
			opts.WebPort = defaultWebPort
		case "--port":
			if i+1 >= len(args) {
				opts.UnknownFlags = append(opts.UnknownFlags, arg)
				continue
			}
			port, err := strconv.Atoi(args[i+1])
			if err != nil {
				opts.UnknownFlags = append(opts.UnknownFlags, arg)
				continue
			}
			opts.WebPort = port
			i++
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
	if o.Help || o.Status || o.Web {
		return nil
	}
	if len(o.UnknownFlags) > 0 {
		return fmt.Errorf("unknown flags provided: %v", o.UnknownFlags)
	}
	return nil
}

func HelpText() string {
	return `Ralph - Autonomous Software Development Agent

Usage:
  ralph                                              # TUI prompt screen (requires a terminal)
  ralph "your feature description"                   # TUI mode
  ralph "your feature description" --dry-run         # Generate PRD only
  ralph --dry-run                                    # Prompt in TUI, then generate PRD only
  ralph --resume                                     # Resume from existing prd.json
  ralph status                                       # Show current PRD status
  ralph web [--port PORT]                            # Start local web UI (default port 8080)

Options:
  --dry-run      Generate PRD only, don't implement
  --resume       Resume implementation from existing prd.json
  --verbose, -v  Enable debug logging
  --help, -h     Show this help message
  --port PORT    Web server port (with ralph web; default 8080)

Environment:
  RALPH_RUNNER   Select the AI runner binary (default: claude; pi, cursor, claude, opencode)
`
}
