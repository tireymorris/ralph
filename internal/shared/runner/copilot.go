package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/constants"
	"ralph/internal/shared/logger"
)

type CopilotRunner struct {
	cfg     *config.Config
	CmdFunc func(ctx context.Context, name string, args ...string) CmdInterface
}

var _ RunnerInterface = (*CopilotRunner)(nil)

func NewCopilot(cfg *config.Config) *CopilotRunner {
	return &CopilotRunner{
		cfg:     cfg,
		CmdFunc: defaultCmdFuncNoStdin(cfg.WorkDir),
	}
}

func (r *CopilotRunner) RunnerName() string {
	return "copilot"
}

func (r *CopilotRunner) CommandName() string {
	return "copilot"
}

func (r *CopilotRunner) IsInternalLog(line string) bool {
	return stderrLineIsInternal(line, stderrFilterDefaultPipedCLI)
}

func (r *CopilotRunner) Run(ctx context.Context, prompt string, outputCh chan<- OutputLine) error {
	args := []string{
		"--allow-all-tools",
		"--allow-all-paths",
		"--no-ask-user",
		"--output-format", "json",
		"--autopilot",
		"--max-autopilot-continues", fmt.Sprintf("%d", constants.CopilotMaxAutopilotContinues),
	}

	logger.Debug("invoking AI runner",
		"runner", r.RunnerName(),
		"command", r.CommandName(),
		"runner", r.cfg.Runner,
		"prompt_length", len(prompt),
		"work_dir", r.cfg.WorkDir)

	if outputCh != nil {
		outputCh <- newStartingOutputLine(r.RunnerName())
	}

	err := runWithPipedCommandAndStdin(ctx, r.CommandName(), r.CmdFunc, strings.NewReader(prompt), args, outputCh,
		parseCopilotJSONL,
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

func parseCopilotJSONL(line string) []OutputLine {
	var event struct {
		Type     string `json:"type"`
		ExitCode int    `json:"exitCode"`
		Data     struct {
			Content      string `json:"content"`
			DeltaContent string `json:"deltaContent"`
			ToolName     string `json:"toolName"`
			Message      string `json:"message"`
			ErrorMessage string `json:"errorMessage"`
			ExitCode     int    `json:"exitCode"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		return []OutputLine{{Text: line, Time: time.Now(), Verbose: true}}
	}

	now := time.Now()

	switch event.Type {
	case "assistant.message_delta":
		if event.Data.DeltaContent != "" {
			return []OutputLine{{Text: event.Data.DeltaContent, Time: now, Append: true}}
		}
	case "tool.execution_start":
		if event.Data.ToolName != "" {
			return []OutputLine{{Text: fmt.Sprintf("Using tool: %s", event.Data.ToolName), Time: now}}
		}
	case "session.error":
		if event.Data.Message != "" {
			return []OutputLine{{Text: event.Data.Message, Time: now, IsErr: true}}
		}
	case "model.call_failure":
		if event.Data.ErrorMessage != "" {
			return []OutputLine{{Text: event.Data.ErrorMessage, Time: now, IsErr: true}}
		}
	case "assistant.message":
		if event.Data.Content != "" {
			return []OutputLine{{Text: event.Data.Content, Time: now, Verbose: true}}
		}
		return []OutputLine{{Text: event.Type, Time: now, Verbose: true}}
	case "assistant.turn_start", "assistant.turn_end", "assistant.message_start":
		return []OutputLine{{Text: event.Type, Time: now, Verbose: true}}
	case "tool.execution_complete", "tool.execution_partial_result":
		text := event.Type
		if event.Data.ToolName != "" {
			text = fmt.Sprintf("%s: %s", event.Type, event.Data.ToolName)
		}
		return []OutputLine{{Text: text, Time: now, Verbose: true}}
	case "result":
		exitCode := copilotResultExitCode(event.ExitCode, event.Data.ExitCode)
		return []OutputLine{{Text: fmt.Sprintf("exit code: %d", exitCode), Time: now, Verbose: true}}
	}

	if strings.HasPrefix(event.Type, "session.") {
		return []OutputLine{{Text: event.Type, Time: now, Verbose: true}}
	}

	return nil
}

func copilotResultExitCode(topLevel, inData int) int {
	// Live Copilot JSONL puts exitCode on the event root; older fixtures nest it under data.
	if topLevel != 0 || inData == 0 {
		return topLevel
	}
	return inData
}
