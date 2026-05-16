package runner

import (
	"testing"

	"ralph/internal/shared/config"
)

func TestRunnerNames(t *testing.T) {
	tests := []struct {
		name        string
		runner      string
		wantRunner  string
		wantCommand string
	}{
		{
			name:        "OpenCode runner returns correct names",
			runner:      "opencode",
			wantRunner:  "OpenCode",
			wantCommand: "opencode",
		},
		{
			name:        "Claude runner returns correct names",
			runner:      "claude",
			wantRunner:  "Claude Code",
			wantCommand: "claude",
		},
		{
			name:        "Another OpenCode runner",
			runner:      "opencode",
			wantRunner:  "OpenCode",
			wantCommand: "opencode",
		},
		{
			name:        "Another Claude runner",
			runner:      "claude",
			wantRunner:  "Claude Code",
			wantCommand: "claude",
		},
		{
			name:        "pi runner",
			runner:      "pi",
			wantRunner:  "pi",
			wantCommand: "pi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Runner:  tt.runner,
				WorkDir: "/tmp",
			}

			runner := New(cfg)

			if runner.RunnerName() != tt.wantRunner {
				t.Errorf("RunnerName() = %q, want %q", runner.RunnerName(), tt.wantRunner)
			}
			if runner.CommandName() != tt.wantCommand {
				t.Errorf("CommandName() = %q, want %q", runner.CommandName(), tt.wantCommand)
			}

			if runner.RunnerName() == "" {
				t.Error("RunnerName() should not be empty")
			}
			if runner.CommandName() == "" {
				t.Error("CommandName() should not be empty")
			}
		})
	}
}
