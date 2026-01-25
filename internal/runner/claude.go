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
	"ralph/internal/constants"
	"ralph/internal/logger"
)

type ClaudeRunner struct {
	cfg     *config.Config
	CmdFunc func(ctx context.Context, name string, args ...string) CmdInterface
}

var _ RunnerInterface = (*ClaudeRunner)(nil)

func (r *ClaudeRunner) RunnerName() string {
	return "Claude Code"
}

func (r *ClaudeRunner) CommandName() string {
	return "claude"
}

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
	modelName := strings.TrimPrefix(r.cfg.Model, "claude-code/")
	if r.cfg.Model != "" {
		args = append(args, "--model", modelName)
	}
	args = append(args, prompt)

	logger.Debug("invoking AI runner",
		"runner", r.RunnerName(),
		"command", r.CommandName(),
		"model", r.cfg.Model,
		"prompt_length", len(prompt),
		"work_dir", r.cfg.WorkDir)

	if outputCh != nil {
		outputCh <- OutputLine{Text: fmt.Sprintf("Starting %s with model %s...", r.RunnerName(), modelName), Time: time.Now()}
	}

	cmd := r.CmdFunc(ctx, r.CommandName(), args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe for %s: %w", r.CommandName(), err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe for %s: %w", r.CommandName(), err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s with model %s: %w", r.CommandName(), modelName, err)
	}

	var wg sync.WaitGroup
	wg.Add(constants.PipeReaderCount)

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
			logger.Debug("AI runner exited with code",
				"runner", r.RunnerName(),
				"command", r.CommandName(),
				"exit_code", exitErr.ExitCode(),
				"model", modelName)
			return fmt.Errorf("%s with model %s exited with code %d", r.RunnerName(), modelName, exitErr.ExitCode())
		}
		return fmt.Errorf("%s with model %s failed: %w", r.RunnerName(), modelName, err)
	}

	logger.Debug("AI runner completed successfully",
		"runner", r.RunnerName(),
		"command", r.CommandName(),
		"model", modelName)
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
