//go:build manual

package runner

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"ralph/internal/shared/config"
)

func manualRunnerPrompt(t *testing.T) string {
	t.Helper()
	path := os.Getenv("MANUAL_RUNNER_PROMPT")
	if path == "" {
		t.Fatal("set MANUAL_RUNNER_PROMPT to a large prompt file path")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read prompt: %v", err)
	}
	return string(b)
}

func TestManualAllRunnersAvoidArgvOverflow(t *testing.T) {
	if os.Getenv("MANUAL_RUNNER_TEST") == "" {
		t.Skip("set MANUAL_RUNNER_TEST=1 to run against real runner binaries")
	}

	prompt := manualRunnerPrompt(t)
	t.Logf("prompt size: %d bytes", len(prompt))

	tests := []struct {
		name   string
		runner string
	}{
		{name: "claude", runner: "claude"},
		{name: "pi", runner: "pi"},
		{name: "cursor-agent", runner: "cursor"},
		{name: "opencode", runner: "opencode"},
		{name: "copilot", runner: "copilot"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{Runner: tc.runner, WorkDir: t.TempDir()}
			r := New(cfg)

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			err := r.Run(ctx, prompt, nil)
			if err != nil && strings.Contains(err.Error(), "argument list too long") {
				t.Fatalf("Run() hit argv overflow: %v", err)
			}
			t.Logf("Run() started without argv overflow (err=%v)", err)
		})
	}
}

func TestManualRunnerCLIsStdinVsArgv(t *testing.T) {
	if os.Getenv("MANUAL_RUNNER_TEST") == "" {
		t.Skip("set MANUAL_RUNNER_TEST=1 to run against real runner binaries")
	}

	prompt := manualRunnerPrompt(t)
	t.Logf("prompt size: %d bytes", len(prompt))

	tests := []struct {
		name   string
		binary string
		flags  []string
	}{
		{
			name:   "claude",
			binary: "claude",
			flags:  []string{"--print", "--output-format", "text"},
		},
		{
			name:   "pi",
			binary: "pi",
			flags:  []string{"--print", "--mode", "json", "--no-session"},
		},
		{
			name:   "cursor-agent",
			binary: "cursor-agent",
			flags:  []string{"--print", "--output-format", "text", "--trust", "--yolo"},
		},
		{
			name:   "opencode",
			binary: "opencode",
			flags:  []string{"run", "--print-logs"},
		},
		{
			name:   "copilot",
			binary: "copilot",
			flags: []string{
				"--allow-all-tools",
				"--allow-all-paths",
				"--no-ask-user",
				"--output-format", "json",
				"--autopilot",
				"--max-autopilot-continues", "50",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name+"/argv_overflows", func(t *testing.T) {
			path, err := exec.LookPath(tc.binary)
			if err != nil {
				t.Skipf("%s not in PATH", tc.binary)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, path, append(append([]string{}, tc.flags...), prompt)...)
			err = cmd.Run()
			if err == nil {
				t.Log("argv path unexpectedly succeeded (ARG_MAX may be high on this machine)")
				return
			}
			if !strings.Contains(err.Error(), "argument list too long") {
				t.Logf("argv path failed differently: %v", err)
				return
			}
			t.Logf("argv path reproduced overflow: %v", err)
		})

		t.Run(tc.name+"/stdin_starts", func(t *testing.T) {
			path, err := exec.LookPath(tc.binary)
			if err != nil {
				t.Skipf("%s not in PATH", tc.binary)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, path, tc.flags...)
			cmd.Stdin = strings.NewReader(prompt)
			err = cmd.Run()
			if err != nil && strings.Contains(err.Error(), "argument list too long") {
				t.Fatalf("stdin path hit argv overflow: %v", err)
			}
			t.Logf("stdin path started without argv overflow (err=%v)", err)
		})
	}
}
