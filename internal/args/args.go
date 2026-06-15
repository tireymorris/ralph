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
	Clean        bool
	Version      bool
	Update       bool
	UpdateRef    string
	UpdateCheck  bool
	Web          bool
	WebPort      int
	SkipCleanup  bool
	AutoApprove  bool
	Headless     bool
	UnknownFlags []string
}

const defaultWebPort = 8080

func Parse(args []string) *Options {
	opts := &Options{}
	var promptParts []string
	inUpdate := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if inUpdate {
			switch arg {
			case "--ref":
				if i+1 >= len(args) {
					opts.UnknownFlags = append(opts.UnknownFlags, arg)
					continue
				}
				opts.UpdateRef = args[i+1]
				i++
				continue
			case "--check":
				opts.UpdateCheck = true
				continue
			default:
				if strings.HasPrefix(arg, "-") {
					opts.UnknownFlags = append(opts.UnknownFlags, arg)
				}
				continue
			}
		}
		switch arg {
		case "--help", "-h":
			opts.Help = true
		case "--dry-run":
			opts.DryRun = true
		case "--resume":
			opts.Resume = true
		case "--verbose", "-v":
			opts.Verbose = true
		case "--skip-cleanup":
			opts.SkipCleanup = true
		case "--yolo":
			opts.AutoApprove = true
		case "--headless":
			opts.Headless = true
			opts.AutoApprove = true
		case "status":
			opts.Status = true
		case "clean":
			opts.Clean = true
		case "version":
			opts.Version = true
		case "update":
			opts.Update = true
			opts.UpdateRef = "main"
			inUpdate = true
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
	if o.AutoApprove {
		switch {
		case o.DryRun:
			return fmt.Errorf("--yolo cannot be used with --dry-run")
		case o.Web:
			return fmt.Errorf("--yolo cannot be used with web")
		case o.Status:
			return fmt.Errorf("--yolo cannot be used with status")
		case o.Clean:
			return fmt.Errorf("--yolo cannot be used with clean")
		case o.Version:
			return fmt.Errorf("--yolo cannot be used with version")
		case o.Update:
			return fmt.Errorf("--yolo cannot be used with update")
		}
	}
	if o.Headless {
		switch {
		case o.DryRun:
			return fmt.Errorf("--headless cannot be used with --dry-run")
		case o.Web:
			return fmt.Errorf("--headless cannot be used with web")
		}
		if !o.Resume && o.Prompt == "" {
			return fmt.Errorf("--headless requires a prompt or --resume")
		}
	}
	if o.Help || o.Status || o.Clean || o.Version || o.Update || o.Web {
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
  ralph --headless "your feature description"        # Unattended yolo mode without the TUI
  ralph "your feature description" --dry-run         # Generate PRD only
  ralph --dry-run                                    # Prompt in TUI, then generate PRD only
  ralph --resume                                     # Resume from existing prd.json
  ralph status                                       # Show current PRD status
  ralph clean                                        # Remove Ralph state files in the working directory
  ralph version                                      # Print build version and commit
  ralph update [--ref REF] [--check]                 # Install or check for updates
  ralph web [--port PORT]                            # Start local web UI (default port 8080)

Options:
  --dry-run        Generate PRD only, don't implement
  --resume         Resume implementation from existing prd.json (--yolo auto-continues without gates)
  --skip-cleanup   Skip post-implementation cleanup phase
  --yolo           Skip manual clarify and PRD approval gates (not with --dry-run or web)
  --headless       Unattended yolo mode without the TUI (--yolo plus no Bubble Tea)
  --verbose, -v    Enable debug logging
  --help, -h       Show this help message
  --port PORT      Web server port (with ralph web; default 8080)
  --ref REF        Git branch or tag for ralph update (default: main)
  --check          With ralph update: compare local commit to remote; exit 2 if update available

Environment:
  RALPH_RUNNER   Select the AI runner binary (default: claude; pi, cursor, claude, opencode, copilot)
  RALPH_YOLO     Set to 1 to skip manual clarify and PRD approval gates
  RALPH_REPO     Git URL for ralph update (default: https://github.com/tireymorris/ralph.git)
`
}
