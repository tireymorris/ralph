package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ralph/internal/config"
	"ralph/internal/shared/logger"
)

type CursorAgentRunner struct {
	cfg     *config.Config
	CmdFunc func(ctx context.Context, name string, args ...string) CmdInterface
}

var _ RunnerInterface = (*CursorAgentRunner)(nil)

func NewCursorAgent(cfg *config.Config) *CursorAgentRunner {
	return &CursorAgentRunner{
		cfg:     cfg,
		CmdFunc: defaultCmdFuncNoStdin(cfg.WorkDir),
	}
}

func (r *CursorAgentRunner) RunnerName() string {
	return "cursor-agent"
}

func (r *CursorAgentRunner) CommandName() string {
	return "cursor-agent"
}

func (r *CursorAgentRunner) IsInternalLog(line string) bool {
	return stderrLineIsInternal(line, stderrFilterDefaultPipedCLI)
}

func (r *CursorAgentRunner) Run(ctx context.Context, prompt string, outputCh chan<- OutputLine) error {
	suffix := strings.TrimPrefix(r.cfg.Model, "cursor-agent/")

	args := []string{"--print", "--output-format", "stream-json", "--trust", "--yolo"}
	if suffix != "" {
		args = append(args, "--model", suffix)
	}
	args = append(args, prompt)

	logger.Debug("invoking AI runner",
		"runner", r.RunnerName(),
		"command", r.CommandName(),
		"model", r.cfg.Model,
		"model_suffix", suffix,
		"prompt_length", len(prompt),
		"work_dir", r.cfg.WorkDir)

	if outputCh != nil {
		outputCh <- newStartingOutputLine(r.RunnerName(), r.cfg.Model)
	}

	err := runWithPipedCommand(ctx, r.CommandName(), r.CmdFunc, args, outputCh,
		parseCursorStreamJSON,
		func(line string) []OutputLine {
			return []OutputLine{{Text: line, IsErr: true, Time: time.Now(), Verbose: r.IsInternalLog(line)}}
		},
	)

	if err != nil {
		logger.Debug("AI runner exited with code",
			"runner", r.RunnerName(),
			"command", r.CommandName(),
			"model", r.cfg.Model)
		return wrapRunnerError(r.RunnerName(), r.cfg.Model, err)
	}

	logger.Debug("AI runner completed successfully",
		"runner", r.RunnerName(),
		"command", r.CommandName(),
		"model", r.cfg.Model,
		"model_suffix", suffix)
	return nil
}

func parseCursorStreamJSON(line string) []OutputLine {
	var event claudeStreamEvent
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		return []OutputLine{{Text: line, Time: time.Now(), Verbose: true}}
	}

	var outputs []OutputLine
	now := time.Now()

	switch event.Type {
	case "assistant":
		for _, content := range event.Message.Content {
			switch content.Type {
			case "text":
				if content.Text != "" {
					outputs = append(outputs, OutputLine{Text: content.Text, Time: now})
				}
			case "tool_use":
				outputs = append(outputs, OutputLine{
					Text: fmt.Sprintf("Using tool: %s", content.Name),
					Time: now,
				})
			}
		}
	case "result":
		if event.Subtype == "success" {
			outputs = append(outputs, OutputLine{Text: "Task completed successfully", Time: now, Verbose: true})
		} else if event.Subtype == "error" {
			outputs = append(outputs, OutputLine{Text: "Task failed", Time: now, IsErr: true})
		}
	}

	return outputs
}
