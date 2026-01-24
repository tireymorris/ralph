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
	"ralph/internal/logger"
)

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

func New(cfg *config.Config) *Runner {
	return &Runner{cfg: cfg, CmdFunc: defaultCmdFunc(cfg.WorkDir)}
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
		outputCh <- OutputLine{Text: "Starting opencode...", Time: time.Now()}
	}

	cmd := r.CmdFunc(ctx, "opencode", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start opencode: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			if outputCh != nil {
				outputCh <- OutputLine{
					Text:    line,
					IsErr:   false,
					Time:    time.Now(),
					Verbose: isVerboseLine(line),
				}
			}
		}
	}()

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			if outputCh != nil {
				outputCh <- OutputLine{
					Text:    line,
					IsErr:   true,
					Time:    time.Now(),
					Verbose: isVerboseLine(line),
				}
			}
		}
	}()

	wg.Wait()
	err = cmd.Wait()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			logger.Debug("opencode exited with code", "exit_code", exitErr.ExitCode())
			return fmt.Errorf("opencode exited with code %d", exitErr.ExitCode())
		}
		return fmt.Errorf("opencode failed: %w", err)
	}

	logger.Debug("opencode completed successfully")
	return nil
}

func isVerboseLine(line string) bool {
	if len(line) >= 4 {
		prefix := line[:4]
		if prefix == "INFO" || prefix == "DEBU" || prefix == "WARN" || prefix == "ERRO" {
			if len(line) > 10 && (strings.Contains(line[:min(30, len(line))], "T") && strings.Contains(line[:min(30, len(line))], ":")) {
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
