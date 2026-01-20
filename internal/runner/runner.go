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
	"ralph/internal/errors"
	"ralph/internal/logger"
)

type OutputLine struct {
	Text    string
	IsErr   bool
	Time    time.Time
	Verbose bool // If true, only show when verbose mode is enabled
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
	args := []string{"run", "--print-logs"}
	if r.cfg.Model != "" {
		args = append(args, "--model", r.cfg.Model)
	}
	// Pass the prompt as a positional argument
	args = append(args, prompt)

	logger.Debug("invoking opencode",
		"model", r.cfg.Model,
		"prompt_length", len(prompt),
		"work_dir", r.cfg.WorkDir)

	if outputCh != nil {
		outputCh <- OutputLine{Text: "Starting opencode...", IsErr: false, Time: time.Now()}
	}

	cmd := r.CmdFunc(ctx, "opencode", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.OpencodeError{Op: "setup", Err: fmt.Errorf("failed to get stdout pipe: %w", err)}
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, errors.OpencodeError{Op: "setup", Err: fmt.Errorf("failed to get stderr pipe: %w", err)}
	}

	if err := cmd.Start(); err != nil {
		return nil, errors.OpencodeError{Op: "setup", Err: fmt.Errorf("failed to start command: %w", err)}
	}

	var outputBuilder strings.Builder
	var wg sync.WaitGroup
	wg.Add(2)

	// Read stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		// Increase buffer size for long lines
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			outputBuilder.WriteString(line + "\n")
			if outputCh != nil {
				outputCh <- OutputLine{
					Text:    line,
					IsErr:   false,
					Time:    time.Now(),
					Verbose: isVerboseLogLine(line),
				}
			}
		}
	}()

	// Read stderr
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
					Verbose: isVerboseLogLine(line),
				}
			}
		}
	}()

	// Wait for both readers to finish before calling Wait()
	wg.Wait()
	err = cmd.Wait()

	result := &Result{
		Output: strings.TrimSpace(outputBuilder.String()),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			logger.Debug("opencode exited with code", "exit_code", result.ExitCode)
		} else {
			result.Error = err
			logger.Debug("opencode error", "error", err)
		}
	} else {
		logger.Debug("opencode completed successfully", "output_length", len(result.Output))
	}

	return result, nil
}

// It handles CSI sequences (\x1b[...X) and OSC sequences (\x1b]...\x07).
// This function is useful for processing terminal output for logging or comparison.
func CleanOutput(output string) string {
	result := output

	// Remove CSI sequences (\x1b[...m, \x1b[...K, etc.)
	for strings.Contains(result, "\x1b[") {
		start := strings.Index(result, "\x1b[")
		end := start + 2
		for end < len(result) && !isCSITerminator(result[end]) {
			end++
		}
		if end < len(result) {
			end++
		}
		result = result[:start] + result[end:]
	}

	// Remove OSC sequences (\x1b]...\x07 or \x1b]...\x1b\\)
	for strings.Contains(result, "\x1b]") {
		start := strings.Index(result, "\x1b]")
		end := start + 2
		// Look for BEL (\x07) or ST (\x1b\\) terminator
		for end < len(result) {
			if result[end] == '\x07' {
				end++
				break
			}
			if end+1 < len(result) && result[end] == '\x1b' && result[end+1] == '\\' {
				end += 2
				break
			}
			end++
		}
		result = result[:start] + result[end:]
	}

	return strings.TrimSpace(result)
}

func isCSITerminator(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

// This filters out noisy service bus messages, internal state updates, etc.
func isVerboseLogLine(line string) bool {
	// Check for structured log format: "INFO|DEBUG|WARN timestamp ..."
	// These are internal opencode logs that are noisy
	if len(line) >= 4 {
		prefix := line[:4]
		if prefix == "INFO" || prefix == "DEBU" || prefix == "WARN" || prefix == "ERRO" {
			// Check if it looks like a structured log line (has timestamp pattern)
			if len(line) > 10 && (strings.Contains(line[:min(30, len(line))], "T") && strings.Contains(line[:min(30, len(line))], ":")) {
				return true
			}
		}
	}

	// Filter service bus and internal messaging logs
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
		" tracking",      // git tracking status lines
		"cwd=/",          // working directory status lines
		"git=/",          // git snapshot status lines
		"stderr=",        // stderr prefix lines
		"Checked ",       // package check lines
		"installed @",    // package install lines
		"[1.00ms]",       // timing lines
		"[2.00ms]",       // timing lines
		"ms] done",       // completion markers like "[2.00ms] done"
		"Saved lockfile", // lockfile messages
	}

	for _, pattern := range verbosePatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}

	return false
}
