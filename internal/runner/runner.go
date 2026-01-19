package runner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"ralph/internal/config"
)

type OutputLine struct {
	Text  string
	IsErr bool
	Time  time.Time
}

type Result struct {
	Output   string
	ExitCode int
	Error    error
}

type CodeRunner interface {
	RunOpenCode(ctx context.Context, prompt string, outputCh chan<- OutputLine) (*Result, error)
}

type Runner struct {
	cfg     *config.Config
	CmdFunc func(ctx context.Context, name string, args ...string) CmdInterface
}

type CmdInterface interface {
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	Start() error
	Wait() error
}

type realCmd struct {
	*exec.Cmd
}

func (c *realCmd) StdinPipe() (io.WriteCloser, error) { return c.Cmd.StdinPipe() }
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

func (r *Runner) RunOpenCode(ctx context.Context, prompt string, outputCh chan<- OutputLine) (*Result, error) {
	args := []string{"run"}
	if r.cfg.Model != "" {
		args = append(args, "--model", r.cfg.Model)
	}

	cmd := r.CmdFunc(ctx, "opencode", args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Channel to signal permission prompts that need auto-approval
	permissionCh := make(chan struct{}, 10)

	// Write initial prompt, then keep stdin open for permission responses
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, prompt)

		// Listen for permission prompts and auto-approve by sending Enter
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-permissionCh:
				if !ok {
					return
				}
				// Send Enter to accept the default "Allow once" option
				io.WriteString(stdin, "\n")
			}
		}
	}()

	var outputBuilder strings.Builder
	doneCh := make(chan struct{})

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			outputBuilder.WriteString(line + "\n")
			if outputCh != nil {
				outputCh <- OutputLine{Text: line, IsErr: false, Time: time.Now()}
			}
			// Detect permission prompts (opencode shows "Permission required:" or "Allow once")
			if strings.Contains(line, "Permission required:") || strings.Contains(line, "● Allow once") {
				select {
				case permissionCh <- struct{}{}:
				default:
				}
			}
		}
		close(permissionCh)
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if outputCh != nil {
				outputCh <- OutputLine{Text: line, IsErr: true, Time: time.Now()}
			}
		}
		close(doneCh)
	}()

	<-doneCh
	err = cmd.Wait()

	result := &Result{
		Output: strings.TrimSpace(outputBuilder.String()),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.Error = err
		}
	}

	return result, nil
}

func CleanOutput(output string) string {
	result := output
	for strings.Contains(result, "\x1b[") {
		start := strings.Index(result, "\x1b[")
		end := start + 2
		for end < len(result) && !isTerminator(result[end]) {
			end++
		}
		if end < len(result) {
			end++
		}
		result = result[:start] + result[end:]
	}
	return strings.TrimSpace(result)
}

func isTerminator(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}
