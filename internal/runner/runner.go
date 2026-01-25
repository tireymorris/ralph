package runner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"ralph/internal/config"
	"ralph/internal/constants"
	"ralph/internal/logger"
)

const (
	// ScannerBufferSize is the maximum line size for reading AI output.
	// Set to 1MB to handle very long output lines.
	ScannerBufferSize = 1024 * 1024

	// PipeReaderCount is the number of goroutines to spawn for reading
	// subprocess stdout and stderr pipes.
	PipeReaderCount = 2
)

type RunnerInterface interface {
	Run(ctx context.Context, prompt string, outputCh chan<- OutputLine) error
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

type CmdInterface interface {
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	Start() error
	Wait() error
}

type realCmd struct {
	*exec.Cmd
}

func (c *realCmd) StdoutPipe() (io.ReadCloser, error) { return c.Cmd.StdoutPipe() }
func (c *realCmd) StderrPipe() (io.ReadCloser, error) { return c.Cmd.StderrPipe() }
func (c *realCmd) Start() error                       { return c.Cmd.Start() }
func (c *realCmd) Wait() error                        { return c.Cmd.Wait() }

func defaultCmdFunc(workDir string) func(ctx context.Context, name string, args ...string) CmdInterface {
	return func(ctx context.Context, name string, args ...string) CmdInterface {
		cmd := exec.CommandContext(ctx, name, args...)
		if workDir != "" {
			cmd.Dir = workDir
		}
		return &realCmd{cmd}
	}
}

func defaultCmdFuncNoStdin(workDir string) func(ctx context.Context, name string, args ...string) CmdInterface {
	return func(ctx context.Context, name string, args ...string) CmdInterface {
		cmd := exec.CommandContext(ctx, name, args...)
		if workDir != "" {
			cmd.Dir = workDir
		}
		cmd.Stdin = nil
		return &realCmd{cmd}
	}
}

type LineTransformer func(line string) []OutputLine

func readPipeLines(pipe io.Reader, outputCh chan<- OutputLine, transform LineTransformer) {
	scanner := bufio.NewScanner(pipe)
	buf := make([]byte, 0, constants.InitialScannerBufferCapacity)
	scanner.Buffer(buf, constants.ScannerBufferSize)
	for scanner.Scan() {
		if outputCh != nil {
			for _, out := range transform(scanner.Text()) {
				outputCh <- out
			}
		}
	}
}

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

func (r *Runner) Run(ctx context.Context, prompt string, outputCh chan<- OutputLine) error {
	args := []string{"run", "--print-logs"}
	if r.cfg.Model != "" {
		args = append(args, "--model", r.cfg.Model)
	}
	args = append(args, prompt)

	logger.Debug("invoking opencode",
		"model", r.cfg.Model,
		"prompt_length", len(prompt),
		"work_dir", r.cfg.WorkDir)

	if outputCh != nil {
		outputCh <- OutputLine{Text: fmt.Sprintf("Starting opencode with model %s...", r.cfg.Model), Time: time.Now()}
	}

	cmd := r.CmdFunc(ctx, "opencode", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe for opencode: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe for opencode: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start opencode with model %s: %w", r.cfg.Model, err)
	}

	var wg sync.WaitGroup
	wg.Add(constants.PipeReaderCount)

	go func() {
		defer wg.Done()
		readPipeLines(stdout, outputCh, func(line string) []OutputLine {
			return []OutputLine{{Text: line, IsErr: false, Time: time.Now(), Verbose: isVerboseLine(line)}}
		})
	}()

	go func() {
		defer wg.Done()
		readPipeLines(stderr, outputCh, func(line string) []OutputLine {
			return []OutputLine{{Text: line, IsErr: true, Time: time.Now(), Verbose: isVerboseLine(line)}}
		})
	}()

	wg.Wait()
	err = cmd.Wait()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			logger.Debug("opencode exited with code", "exit_code", exitErr.ExitCode(), "model", r.cfg.Model)
			return fmt.Errorf("opencode with model %s exited with code %d", r.cfg.Model, exitErr.ExitCode())
		}
		return fmt.Errorf("opencode with model %s failed: %w", r.cfg.Model, err)
	}

	logger.Debug("opencode completed successfully", "model", r.cfg.Model)
	return nil
}

func isVerboseLine(line string) bool {
	if len(line) >= constants.VerbosePatternMinLength {
		prefix := line[:constants.VerbosePatternMinLength]
		if prefix == "INFO" || prefix == "DEBU" || prefix == "WARN" || prefix == "ERRO" {
			if len(line) > constants.VerboseTimestampMinLength && (strings.Contains(line[:min(constants.TimestampContextLength, len(line))], "T") && strings.Contains(line[:min(constants.TimestampContextLength, len(line))], ":")) {
				return true
			}
		}
	}

	verbosePatterns := []string{
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

	for _, pattern := range verbosePatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}

	return false
}
