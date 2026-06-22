package app

import (
	"context"
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
	"ralph/internal/update"
	"ralph/internal/version"
	"ralph/internal/web"
)

type Coordinator struct {
	loadConfig     func() (*config.Config, error)
	runClean       func(*config.Config) int
	runStatus      func(*config.Config) int
	runTUI         func(*config.Config, string, bool, bool, bool) int
	runHeadless    func(*config.Config, string, bool) int
	runUpdate      func(*args.Options) int
	runUpdateCheck func(*args.Options) int
	runWeb         func(*config.Config, int) int
	validateGit    func(string) error
	validateResume func(*config.Config, bool) error
	helpText       func() string
	versionInfo    func() string
	isTerminal     func(fd uintptr) bool
}

func newCoordinator() *Coordinator {
	return &Coordinator{
		loadConfig:     config.Load,
		runClean:       runClean,
		runStatus:      runStatus,
		runTUI:         runTUI,
		runHeadless:    runHeadless,
		runUpdate:      RunUpdate,
		runUpdateCheck: RunUpdateCheck,
		runWeb:         runWeb,
		validateGit:    workdir.ValidateGit,
		validateResume: validateResume,
		helpText:       args.HelpText,
		versionInfo:    version.Info,
		isTerminal:     isatty.IsTerminal,
	}
}

func (c *Coordinator) Run(opts *args.Options) int {
	c = c.withDefaults()

	if opts.Help {
		fmt.Print(c.helpText())
		return 0
	}
	if err := opts.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Print(c.helpText())
		return 1
	}
	for _, flag := range opts.UnknownFlags {
		fmt.Fprintf(os.Stderr, "Warning: unknown flag %q (ignored)\n", flag)
	}
	if opts.Version {
		fmt.Println(c.versionInfo())
		return 0
	}
	if opts.Update {
		if opts.UpdateCheck {
			return c.runUpdateCheck(opts)
		}
		return c.runUpdate(opts)
	}

	logger.Init(opts.Verbose)
	logger.Debug("starting ralph", "verbose", opts.Verbose)

	cfg, err := c.loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		fmt.Print(c.helpText())
		return 1
	}
	logger.Debug("config loaded", "runner", cfg.Runner)

	if opts.Clean {
		return c.runClean(cfg)
	}

	if err := c.validateResume(cfg, opts.Resume); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	applyRuntimeOptions(cfg, opts)

	if opts.Status {
		return c.runStatus(cfg)
	}
	if opts.Web {
		return c.runWeb(cfg, opts.WebPort)
	}
	if opts.Headless {
		if opts.Prompt != "" || opts.Resume {
			if err := c.validateGit(cfg.WorkDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return 1
			}
		}
		return c.runHeadless(cfg, opts.Prompt, opts.Resume)
	}
	if opts.Prompt == "" && !opts.Resume && !c.isTerminal(os.Stdin.Fd()) {
		fmt.Fprintf(os.Stderr, "Error: interactive prompt requires a terminal (provide a prompt argument or use --resume)\n")
		return 1
	}
	if opts.Prompt != "" || opts.Resume {
		if err := c.validateGit(cfg.WorkDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	}
	attemptBootUpdate(opts)
	return c.runTUI(cfg, opts.Prompt, opts.DryRun, opts.Resume, opts.Verbose)
}

func attemptBootUpdate(opts *args.Options) {
	repo := update.RepoFromEnv()
	ref := opts.UpdateRef
	if ref == "" {
		ref = update.DefaultRef
	}
	_, _ = attemptUpdate(context.Background(), repo, ref)
}

func (c *Coordinator) withDefaults() *Coordinator {
	if c == nil {
		c = newCoordinator()
	}
	if c.loadConfig == nil {
		c.loadConfig = config.Load
	}
	if c.runClean == nil {
		c.runClean = runClean
	}
	if c.runStatus == nil {
		c.runStatus = runStatus
	}
	if c.runTUI == nil {
		c.runTUI = runTUI
	}
	if c.runHeadless == nil {
		c.runHeadless = runHeadless
	}
	if c.runUpdate == nil {
		c.runUpdate = RunUpdate
	}
	if c.runUpdateCheck == nil {
		c.runUpdateCheck = RunUpdateCheck
	}
	if c.runWeb == nil {
		c.runWeb = runWeb
	}
	if c.validateGit == nil {
		c.validateGit = workdir.ValidateGit
	}
	if c.validateResume == nil {
		c.validateResume = validateResume
	}
	if c.helpText == nil {
		c.helpText = args.HelpText
	}
	if c.versionInfo == nil {
		c.versionInfo = version.Info
	}
	if c.isTerminal == nil {
		c.isTerminal = isatty.IsTerminal
	}
	return c
}

func applyRuntimeOptions(cfg *config.Config, opts *args.Options) {
	cfg.SkipCleanup = opts.SkipCleanup
	cfg.DryRun = opts.DryRun
	cfg.AutoApprove = opts.AutoApprove || cfg.AutoApprove
}

func runTUI(cfg *config.Config, prompt string, dryRun, resume, verbose bool) int {
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

func runWeb(cfg *config.Config, port int) int {
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

func runStatus(cfg *config.Config) int {
	if err := status.Display(cfg); err != nil {
		return 1
	}
	return 0
}

func runClean(cfg *config.Config) int {
	if err := clean.RemoveState(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Println("Ralph state removed.")
	return 0
}

func validateResume(cfg *config.Config, resume bool) error {
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

func RunWeb(cfg *config.Config, port int) int {
	return runWeb(cfg, port)
}

func ValidateResume(cfg *config.Config, resume bool) error {
	return validateResume(cfg, resume)
}
