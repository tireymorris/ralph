package runner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/logger"
)

type RunnerInterface interface {
	Run(ctx context.Context, prompt string, outputCh chan<- OutputLine) error
	RunnerName() string
	CommandName() string
	IsInternalLog(line string) bool
}

type OutputLine struct {
	Text    string
	IsErr   bool
	Time    time.Time
	Verbose bool
}

type Runner struct {
	cfg     *config.Config
	CmdFunc func(ctx context.Context, name string, args ...string) CmdInterface
}

var _ RunnerInterface = (*Runner)(nil)

func New(cfg *config.Config) RunnerInterface {
	provider := config.DetectRunner(cfg.Runner)
	if provider == config.RunnerClaude {
		logger.Debug("using Claude Code runner", "runner", cfg.Runner)
		return NewClaude(cfg)
	}
	if provider == config.RunnerPi {
		logger.Debug("using pi runner", "runner", cfg.Runner)
		return NewPi(cfg)
	}
	if provider == config.RunnerCursor {
		logger.Debug("using cursor-agent runner", "runner", cfg.Runner)
		return NewCursorAgent(cfg)
	}
	if provider == config.RunnerMock {
		logger.Debug("using mock runner", "runner", cfg.Runner)
		return NewMock(cfg)
	}

	logger.Debug("using OpenCode runner", "runner", cfg.Runner)
	return &Runner{cfg: cfg, CmdFunc: defaultCmdFunc(cfg.WorkDir)}
}

func NewWithError(cfg *config.Config) (RunnerInterface, error) {
	if err := cfg.ValidateRunner(); err != nil {
		return nil, fmt.Errorf("invalid runner configuration %q: %w", cfg.Runner, err)
	}

	return New(cfg), nil
}

func (r *Runner) RunnerName() string {
	return "OpenCode"
}

func (r *Runner) CommandName() string {
	return "opencode"
}

func (r *Runner) IsInternalLog(line string) bool {
	return stderrLineIsInternal(line, stderrFilterOpenCode)
}

func (r *Runner) Run(ctx context.Context, prompt string, outputCh chan<- OutputLine) error {
	args := []string{"run", "--print-logs", prompt}

	logger.Debug("invoking AI runner",
		"runner", r.RunnerName(),
		"command", r.CommandName(),
		"runner", r.cfg.Runner,
		"prompt_length", len(prompt),
		"work_dir", r.cfg.WorkDir)

	if outputCh != nil {
		outputCh <- OutputLine{Text: fmt.Sprintf("Starting %s...", r.RunnerName()), Time: time.Now()}
	}

	cmd := r.CmdFunc(ctx, r.CommandName(), args...)
	err := runPipedCommand(r.CommandName(), cmd, outputCh,
		func(line string) []OutputLine {
			return []OutputLine{{Text: line, IsErr: false, Time: time.Now(), Verbose: r.IsInternalLog(line)}}
		},
		func(line string) []OutputLine {
			return []OutputLine{{Text: line, IsErr: true, Time: time.Now(), Verbose: r.IsInternalLog(line)}}
		},
	)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			logger.Debug("AI runner exited with code",
				"runner", r.RunnerName(),
				"command", r.CommandName(),
				"exit_code", exitErr.ExitCode(),
				"runner", r.cfg.Runner)
			return fmt.Errorf("%s exited with code %d", r.RunnerName(), exitErr.ExitCode())
		}
		return fmt.Errorf("%s failed: %w", r.RunnerName(), err)
	}

	logger.Debug("AI runner completed successfully",
		"runner", r.RunnerName(),
		"command", r.CommandName(),
		"runner", r.cfg.Runner)
	return nil
}

// isOpenCodeInternalLog identifies OpenCode noise that should be hidden unless verbose.
func isOpenCodeInternalLog(line string) bool {
	if len(line) >= constants.VerbosePatternMinLength {
		prefix := line[:constants.VerbosePatternMinLength]
		if prefix == "INFO" || prefix == "DEBU" || prefix == "WARN" || prefix == "ERRO" {
			if len(line) > constants.VerboseTimestampMinLength && (strings.Contains(line[:min(constants.TimestampContextLength, len(line))], "T") && strings.Contains(line[:min(constants.TimestampContextLength, len(line))], ":")) {
				return true
			}
		}
	}

	internalLogPatterns := []string{
		"service=bus",
		"type=message.",
		"publishing",
		"subscribing",
		"service=provider",
		"service=session",
		"service=lsp",
		"service=file",
		"service=default",
		" tracking",
		"cwd=/",
		"git=/",
		"stderr=",
		"Checked ",
		"installed @",
		"[1.00ms]",
		"[2.00ms]",
		"ms] done",
		"Saved lockfile",
	}

	for _, pattern := range internalLogPatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}

	return false
}

type stderrFilter int

const (
	stderrFilterOpenCode stderrFilter = iota
	stderrFilterDefaultPipedCLI
)

func stderrLineIsInternal(line string, f stderrFilter) bool {
	if isUserError(line) {
		return false
	}
	if f == stderrFilterOpenCode {
		return isOpenCodeInternalLog(line)
	}
	return true
}

func isUserError(line string) bool {
	userErrorPatterns := []string{
		"error:",
		"failed:",
		"cannot:",
		"unable:",
		"permission denied",
		"file not found",
		"no such file",
		"invalid",
		"error",
		"failed",
		"cannot",
		"unable",
	}

	lowerLine := strings.ToLower(line)
	for _, pattern := range userErrorPatterns {
		if strings.Contains(lowerLine, pattern) {
			return true
		}
	}

	return false
}
