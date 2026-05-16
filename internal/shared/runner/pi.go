package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/logger"
)

type PiRunner struct {
	cfg     *config.Config
	CmdFunc func(ctx context.Context, name string, args ...string) CmdInterface
}

var _ RunnerInterface = (*PiRunner)(nil)

func NewPi(cfg *config.Config) *PiRunner {
	return &PiRunner{
		cfg:     cfg,
		CmdFunc: defaultCmdFuncNoStdin(cfg.WorkDir),
	}
}

func (r *PiRunner) RunnerName() string {
	return "pi"
}

func (r *PiRunner) CommandName() string {
	return "pi"
}

func (r *PiRunner) IsInternalLog(line string) bool {
	return stderrLineIsInternal(line, stderrFilterDefaultPipedCLI)
}

func (r *PiRunner) Run(ctx context.Context, prompt string, outputCh chan<- OutputLine) error {
	args := []string{"--print", "--mode", "json", "--no-session", prompt}

	logger.Debug("invoking AI runner",
		"runner", r.RunnerName(),
		"command", r.CommandName(),
		"runner", r.cfg.Runner,
		"prompt_length", len(prompt),
		"work_dir", r.cfg.WorkDir)

	if outputCh != nil {
		outputCh <- newStartingOutputLine(r.RunnerName())
	}

	err := runWithPipedCommand(ctx, r.CommandName(), r.CmdFunc, args, outputCh,
		parsePiJSONLine,
		func(line string) []OutputLine {
			return []OutputLine{{Text: line, IsErr: true, Time: time.Now(), Verbose: r.IsInternalLog(line)}}
		},
	)

	if err != nil {
		logger.Debug("AI runner exited with code",
			"runner", r.RunnerName(),
			"command", r.CommandName(),
			"runner", r.cfg.Runner)
		return wrapRunnerError(r.RunnerName(), err)
	}

	logger.Debug("AI runner completed successfully",
		"runner", r.RunnerName(),
		"command", r.CommandName(),
		"runner", r.cfg.Runner)
	return nil
}

type piAssistantMsgEvt struct {
	Type  string `json:"type"`
	Delta string `json:"delta"`
}

func parsePiJSONLine(line string) []OutputLine {
	now := time.Now()
	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(line), &head); err != nil {
		return []OutputLine{{Text: line, Time: now, Verbose: true}}
	}

	switch head.Type {
	case "message_update":
		var envelope struct {
			AssistantMessageEvent json.RawMessage `json:"assistantMessageEvent"`
		}
		if err := json.Unmarshal([]byte(line), &envelope); err != nil {
			return []OutputLine{{Text: line, Time: now, Verbose: true}}
		}
		if len(envelope.AssistantMessageEvent) == 0 {
			return nil
		}
		var am piAssistantMsgEvt
		if err := json.Unmarshal(envelope.AssistantMessageEvent, &am); err != nil {
			return nil
		}
		if am.Type == "text_delta" && am.Delta != "" {
			return []OutputLine{{Text: am.Delta, Time: now}}
		}
	case "tool_execution_start":
		var ts struct {
			ToolName string `json:"toolName"`
		}
		if err := json.Unmarshal([]byte(line), &ts); err != nil {
			return nil
		}
		if ts.ToolName != "" {
			return []OutputLine{{Text: fmt.Sprintf("Using tool: %s", ts.ToolName), Time: now}}
		}
	case "tool_execution_end":
		var te struct {
			ToolName string `json:"toolName"`
			IsError  bool   `json:"isError"`
		}
		if err := json.Unmarshal([]byte(line), &te); err != nil {
			return nil
		}
		if te.IsError {
			return []OutputLine{{Text: fmt.Sprintf("Tool %s failed", te.ToolName), Time: now, IsErr: true}}
		}
	}
	return nil
}
