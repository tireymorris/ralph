package runner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"ralph/internal/config"
	"ralph/internal/constants"
	"ralph/internal/logger"
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

func isClaudeCodeModel(model string) bool {
	return strings.HasPrefix(model, "claude-code/")
}

func New(cfg *config.Config) RunnerInterface {
	if isClaudeCodeModel(cfg.Model) {
		logger.Debug("using Claude Code runner", "model", cfg.Model)
		return NewClaude(cfg)
	}

	logger.Debug("using OpenCode runner", "model", cfg.Model)
	return &Runner{cfg: cfg, CmdFunc: defaultCmdFunc(cfg.WorkDir)}
}

func NewWithError(cfg *config.Config) (RunnerInterface, error) {
	if err := cfg.ValidateModel(); err != nil {
		return nil, fmt.Errorf("invalid model configuration %q: %w", cfg.Model, err)
	}

	return New(cfg), nil
}

func (r *Runner) RunnerName() string {
	return "OpenCode"
}

func (r *Runner) CommandName() string {
	return "opencode"
}

// IsInternalLog determines if a line contains OpenCode-specific internal log patterns
// that should be filtered out as verbose noise rather than shown to the user.
func (r *Runner) IsInternalLog(line string) bool {
	return isOpenCodeInternalLog(line)
}

func (r *Runner) Run(ctx context.Context, prompt string, outputCh chan<- OutputLine) error {
	args := []string{"run", "--print-logs"}
	if r.cfg.Model != "" {
		args = append(args, "--model", r.cfg.Model)
	}
	args = append(args, prompt)

	logger.Debug("invoking AI runner",
		"runner", r.RunnerName(),
		"command", r.CommandName(),
		"model", r.cfg.Model,
		"prompt_length", len(prompt),
		"work_dir", r.cfg.WorkDir)

	if outputCh != nil {
		outputCh <- OutputLine{Text: fmt.Sprintf("Starting %s with model %s...", r.RunnerName(), r.cfg.Model), Time: time.Now()}
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
				"model", r.cfg.Model)
			return fmt.Errorf("%s with model %s exited with code %d", r.RunnerName(), r.cfg.Model, exitErr.ExitCode())
		}
		return fmt.Errorf("%s with model %s failed: %w", r.RunnerName(), r.cfg.Model, err)
	}

	logger.Debug("AI runner completed successfully",
		"runner", r.RunnerName(),
		"command", r.CommandName(),
		"model", r.cfg.Model)
	return nil
}

// isOpenCodeInternalLog determines if a line contains OpenCode-specific internal log patterns
// that should be filtered out as verbose noise rather than shown to the user.
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
