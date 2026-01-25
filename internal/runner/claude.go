package runner

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"ralph/internal/config"
	"ralph/internal/logger"
)

type ClaudeRunner struct {
	cfg     *config.Config
	CmdFunc func(ctx context.Context, name string, args ...string) CmdInterface
}

var _ RunnerInterface = (*ClaudeRunner)(nil)

func NewClaude(cfg *config.Config) *ClaudeRunner {
	return &ClaudeRunner{
		cfg:     cfg,
		CmdFunc: defaultCmdFunc(cfg.WorkDir),
	}
}

func (r *ClaudeRunner) Run(ctx context.Context, prompt string, outputCh chan<- OutputLine) error {
	args := []string{
		"--print",
		"--dangerously-skip-permissions",
	}
	if r.cfg.Model != "" {
		args = append(args, "--model", strings.TrimPrefix(r.cfg.Model, "claude-code/"))
	}
	args = append(args, prompt)

	logger.Debug("invoking claude",
		"model", r.cfg.Model,
		"prompt_length", len(prompt),
		"work_dir", r.cfg.WorkDir)

	if outputCh != nil {
		outputCh <- OutputLine{Text: "Starting claude...", Time: time.Now()}
	}

	cmd := r.CmdFunc(ctx, "claude", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start claude: %w", err)
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
					Verbose: isClaudeVerboseLine(line),
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
					Verbose: isClaudeVerboseLine(line),
				}
			}
		}
	}()

	wg.Wait()
	err = cmd.Wait()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			logger.Debug("claude exited with code", "exit_code", exitErr.ExitCode())
			return fmt.Errorf("claude exited with code %d", exitErr.ExitCode())
		}
		return fmt.Errorf("claude failed: %w", err)
	}

	logger.Debug("claude completed successfully")
	return nil
}

func isClaudeVerboseLine(line string) bool {
	if len(line) == 0 {
		return false
	}

	verbosePatterns := []string{
		"[DEBUG]",
		"[INFO]",
		"[WARN]",
		"[ERROR]",
		"TRACE:",
		"Tool execution:",
		"API request:",
		"API response:",
		"Token usage:",
		"Process ID:",
		"Working directory:",
		"Git repository:",
		"Session ID:",
		"Model:",
		"Temperature:",
		"Max tokens:",
		"duration=",
		"status=",
		"bytes=",
		"files=",
		"request_id=",
		"timestamp=",
	}

	for _, pattern := range verbosePatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}

	if strings.HasPrefix(line, " ") && len(line) > 1 {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "|") || strings.HasPrefix(trimmed, "└") || strings.HasPrefix(trimmed, "├") {
			return true
		}
	}

	return false
}
