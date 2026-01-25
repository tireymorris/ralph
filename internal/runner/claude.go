package runner

import (
	"context"
	"encoding/json"
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
		CmdFunc: defaultCmdFuncNoStdin(cfg.WorkDir),
	}
}

func (r *ClaudeRunner) Run(ctx context.Context, prompt string, outputCh chan<- OutputLine) error {
	args := []string{
		"--print",
		"--verbose",
		"--output-format", "stream-json",
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
		readPipeLines(stdout, outputCh, parseClaudeStreamJSON)
	}()

	go func() {
		defer wg.Done()
		readPipeLines(stderr, outputCh, func(line string) []OutputLine {
			return []OutputLine{{Text: line, IsErr: true, Time: time.Now(), Verbose: true}}
		})
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

type claudeStreamEvent struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype,omitempty"`
	Message struct {
		Content []struct {
			Type  string `json:"type"`
			Text  string `json:"text,omitempty"`
			Name  string `json:"name,omitempty"`
			Input any    `json:"input,omitempty"`
		} `json:"content"`
	} `json:"message,omitempty"`
	Result string `json:"result,omitempty"`
}

func parseClaudeStreamJSON(line string) []OutputLine {
	var event claudeStreamEvent
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		return []OutputLine{{Text: line, Time: time.Now(), Verbose: true}}
	}

	var outputs []OutputLine
	now := time.Now()

	switch event.Type {
	case "system":
		if event.Subtype == "init" {
			outputs = append(outputs, OutputLine{Text: "Claude initialized", Time: now, Verbose: true})
		}
	case "assistant":
		for _, content := range event.Message.Content {
			switch content.Type {
			case "text":
				if content.Text != "" {
					outputs = append(outputs, OutputLine{Text: content.Text, Time: now})
				}
			case "tool_use":
				outputs = append(outputs, OutputLine{
					Text:    fmt.Sprintf("Using tool: %s", content.Name),
					Time:    now,
					Verbose: false,
				})
			}
		}
	case "user":
		outputs = append(outputs, OutputLine{Text: "Tool completed", Time: now, Verbose: true})
	case "result":
		if event.Subtype == "success" {
			outputs = append(outputs, OutputLine{Text: "Task completed successfully", Time: now, Verbose: true})
		} else if event.Subtype == "error" {
			outputs = append(outputs, OutputLine{Text: "Task failed", Time: now, IsErr: true})
		}
	}

	return outputs
}
