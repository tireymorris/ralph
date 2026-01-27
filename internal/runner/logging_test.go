package runner

import (
	"testing"

	"ralph/internal/config"
)

func TestRunnerNames(t *testing.T) {
	tests := []struct {
		name        string
		model       string
		wantRunner  string
		wantCommand string
	}{
		{
			name:        "OpenCode runner returns correct names",
			model:       "opencode/big-pickle",
			wantRunner:  "OpenCode",
			wantCommand: "opencode",
		},
		{
			name:        "Claude runner returns correct names",
			model:       "claude-code/claude-3.5-sonnet",
			wantRunner:  "Claude Code",
			wantCommand: "claude",
		},
		{
			name:        "Another OpenCode model",
			model:       "opencode/big-pickle",
			wantRunner:  "OpenCode",
			wantCommand: "opencode",
		},
		{
			name:        "Another Claude model",
			model:       "claude-code/claude-3.5-haiku",
			wantRunner:  "Claude Code",
			wantCommand: "claude",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Model:   tt.model,
				WorkDir: "/tmp",
			}

			var runner RunnerInterface
			if isClaudeCodeModel(tt.model) {
				runner = NewClaude(cfg)
			} else {
				runner = &Runner{cfg: cfg, CmdFunc: defaultCmdFunc(cfg.WorkDir)}
			}

			// Verify runner methods return correct names
			if runner.RunnerName() != tt.wantRunner {
				t.Errorf("RunnerName() = %q, want %q", runner.RunnerName(), tt.wantRunner)
			}
			if runner.CommandName() != tt.wantCommand {
				t.Errorf("CommandName() = %q, want %q", runner.CommandName(), tt.wantCommand)
			}

			// Verify that methods return non-empty values
			if runner.RunnerName() == "" {
				t.Error("RunnerName() should not be empty")
			}
			if runner.CommandName() == "" {
				t.Error("CommandName() should not be empty")
			}
		})
	}
}
