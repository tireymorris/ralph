package app

import (
	"fmt"
	"net/http"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"

	"ralph/internal/args"
	"ralph/internal/clean"
	"ralph/internal/shared/config"
	"ralph/internal/shared/logger"
	sharedprd "ralph/internal/shared/prd"
	"ralph/internal/shared/workdir"
	"ralph/internal/status"
	"ralph/internal/tui"
	"ralph/internal/version"
	"ralph/internal/web"
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
	if opts.Version {
		fmt.Println(version.Info())
		return 0
	}
	if opts.Update {
		if opts.UpdateCheck {
			return RunUpdateCheck(opts)
		}
		return RunUpdate(opts)
	}
	logger.Init(opts.Verbose)
	logger.Debug("starting ralph", "verbose", opts.Verbose)

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		fmt.Print(args.HelpText())
		return 1
	}
	logger.Debug("config loaded", "runner", cfg.Runner)

	if opts.Clean {
		return RunClean(cfg)
	}

	if err := ValidateResume(cfg, opts.Resume); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	applyRuntimeOptions(cfg, opts)

	if opts.Status {
		return RunStatus(cfg)
	}
	if opts.Web {
		return RunWeb(cfg, opts.WebPort)
	}
	if opts.Prompt == "" && !opts.Resume && !isatty.IsTerminal(os.Stdin.Fd()) {
		fmt.Fprintf(os.Stderr, "Error: interactive prompt requires a terminal (provide a prompt argument or use --resume)\n")
		return 1
	}
	if opts.Prompt != "" || opts.Resume {
		if err := workdir.ValidateGit(cfg.WorkDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	}
	return RunTUI(cfg, opts.Prompt, opts.DryRun, opts.Resume, opts.Verbose)
}

func applyRuntimeOptions(cfg *config.Config, opts *args.Options) {
	cfg.SkipCleanup = opts.SkipCleanup
	cfg.DryRun = opts.DryRun
	cfg.AutoApprove = opts.AutoApprove || cfg.AutoApprove
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

func RunWeb(cfg *config.Config, port int) int {
	addr := web.ListenAddr(port)
	errCh := make(chan error, 1)
	go func() {
		errCh <- web.Run(cfg, addr)
	}()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if web.ServerURL() != "" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if u := web.ServerURL(); u != "" {
		fmt.Fprintf(os.Stdout, "ralph web listening on %s\n", u)
	}

	err := <-errCh
	if err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "Error running web server: %v\n", err)
		return 1
	}
	return 0
}

func RunStatus(cfg *config.Config) int {
	if err := status.Display(cfg); err != nil {
		return 1
	}
	return 0
}

func RunClean(cfg *config.Config) int {
	if err := clean.RemoveState(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Println("Ralph state removed.")
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
