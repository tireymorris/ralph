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

// OutputLine represents a line of output from a command
type OutputLine struct {
	Text  string
	IsErr bool
	Time  time.Time
}

// Result represents the result of a command execution
type Result struct {
	Output   string
	ExitCode int
	Error    error
}

// Runner executes commands
type Runner struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Runner {
	return &Runner{cfg: cfg}
}

// RunOpenCode runs the opencode command with the given prompt
func (r *Runner) RunOpenCode(ctx context.Context, prompt string, outputCh chan<- OutputLine) (*Result, error) {
	args := []string{"run"}
	if r.cfg.Model != "" {
		args = append(args, "--model", r.cfg.Model)
	}

	cmd := exec.CommandContext(ctx, "opencode", args...)

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

	// Write prompt to stdin
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, prompt)
	}()

	// Collect output
	var outputBuilder strings.Builder
	doneCh := make(chan struct{})

	// Read stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			outputBuilder.WriteString(line + "\n")
			if outputCh != nil {
				outputCh <- OutputLine{Text: line, IsErr: false, Time: time.Now()}
			}
		}
	}()

	// Read stderr
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

// CleanOutput removes ANSI escape codes and normalizes whitespace
func CleanOutput(output string) string {
	// Remove ANSI escape codes (simplified)
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
