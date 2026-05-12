package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ralph/internal/config"
	"ralph/internal/logger"
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

func piProviderAndModel(cfgModel string) (provider, piModel string) {
	rest := strings.TrimPrefix(cfgModel, "pi/")
	if rest == "" {
		return "", ""
	}
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) == 1 {
		return "", parts[0]
	}
	p := parts[0]
	m := parts[1]
	if p == "" {
		return "", m
	}
	return p, m
}

func (r *PiRunner) Run(ctx context.Context, prompt string, outputCh chan<- OutputLine) error {
	provider, piModel := piProviderAndModel(r.cfg.Model)
	args := []string{"--print", "--mode", "json", "--no-session"}
	if provider != "" {
		args = append(args, "--provider", provider)
	}
	if piModel != "" {
		args = append(args, "--model", piModel)
	}
	args = append(args, prompt)

	logger.Debug("invoking AI runner",
		"runner", r.RunnerName(),
		"command", r.CommandName(),
		"model", r.cfg.Model,
		"pi_provider", provider,
		"pi_model", piModel,
		"prompt_length", len(prompt),
		"work_dir", r.cfg.WorkDir)

	if outputCh != nil {
		outputCh <- newStartingOutputLine(r.RunnerName(), r.cfg.Model)
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
			"model", r.cfg.Model,
			"pi_provider", provider,
			"pi_model", piModel)
		return wrapRunnerError(r.RunnerName(), r.cfg.Model, err)
	}

	logger.Debug("AI runner completed successfully",
		"runner", r.RunnerName(),
		"command", r.CommandName(),
		"model", r.cfg.Model,
		"pi_provider", provider,
		"pi_model", piModel)
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
