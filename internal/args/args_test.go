package args

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Options
	}{
		{name: "empty args", args: []string{}, expected: Options{}},
		{name: "help flag short", args: []string{"-h"}, expected: Options{Help: true}},
		{name: "help flag long", args: []string{"--help"}, expected: Options{Help: true}},
		{name: "dry run flag", args: []string{"--dry-run"}, expected: Options{DryRun: true}},
		{name: "yolo flag", args: []string{"--yolo"}, expected: Options{AutoApprove: true}},
		{name: "resume flag", args: []string{"--resume"}, expected: Options{Resume: true}},
		{name: "verbose flag short", args: []string{"-v"}, expected: Options{Verbose: true}},
		{name: "verbose flag long", args: []string{"--verbose"}, expected: Options{Verbose: true}},
		{name: "single prompt word", args: []string{"hello"}, expected: Options{Prompt: "hello"}},
		{name: "multi word prompt", args: []string{"hello", "world"}, expected: Options{Prompt: "hello world"}},
		{name: "prompt with flags", args: []string{"Add", "feature", "--dry-run"}, expected: Options{Prompt: "Add feature", DryRun: true}},
		{name: "unknown flag captured", args: []string{"--unknown", "prompt"}, expected: Options{Prompt: "prompt", UnknownFlags: []string{"--unknown"}}},
		{name: "multiple unknown flags captured", args: []string{"--foo", "-x", "prompt", "--bar"}, expected: Options{Prompt: "prompt", UnknownFlags: []string{"--foo", "-x", "--bar"}}},
		{name: "status command", args: []string{"status"}, expected: Options{Status: true}},
		{name: "clean command", args: []string{"clean"}, expected: Options{Clean: true}},
		{name: "version command", args: []string{"version"}, expected: Options{Version: true}},
		{name: "update command", args: []string{"update"}, expected: Options{Update: true, UpdateRef: "main"}},
		{name: "update with ref", args: []string{"update", "--ref", "v1.0"}, expected: Options{Update: true, UpdateRef: "v1.0"}},
		{name: "update check", args: []string{"update", "--check"}, expected: Options{Update: true, UpdateRef: "main", UpdateCheck: true}},
		{name: "update with yolo is unknown", args: []string{"update", "--yolo"}, expected: Options{Update: true, UpdateRef: "main", UnknownFlags: []string{"--yolo"}}},
		{name: "web command", args: []string{"web"}, expected: Options{Web: true, WebPort: 8080}},
		{name: "web with port", args: []string{"web", "--port", "3000"}, expected: Options{Web: true, WebPort: 3000}},
		{name: "skip cleanup flag", args: []string{"--skip-cleanup", "do thing"}, expected: Options{Prompt: "do thing", SkipCleanup: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.args)
			if got.Status != tt.expected.Status {
				t.Errorf("Status = %v, want %v", got.Status, tt.expected.Status)
			}
			if got.Clean != tt.expected.Clean {
				t.Errorf("Clean = %v, want %v", got.Clean, tt.expected.Clean)
			}
			if got.Version != tt.expected.Version {
				t.Errorf("Version = %v, want %v", got.Version, tt.expected.Version)
			}
			if got.Update != tt.expected.Update {
				t.Errorf("Update = %v, want %v", got.Update, tt.expected.Update)
			}
			if got.UpdateRef != tt.expected.UpdateRef {
				t.Errorf("UpdateRef = %q, want %q", got.UpdateRef, tt.expected.UpdateRef)
			}
			if got.UpdateCheck != tt.expected.UpdateCheck {
				t.Errorf("UpdateCheck = %v, want %v", got.UpdateCheck, tt.expected.UpdateCheck)
			}
			if got.Prompt != tt.expected.Prompt {
				t.Errorf("Prompt = %q, want %q", got.Prompt, tt.expected.Prompt)
			}
			if got.DryRun != tt.expected.DryRun {
				t.Errorf("DryRun = %v, want %v", got.DryRun, tt.expected.DryRun)
			}
			if got.Resume != tt.expected.Resume {
				t.Errorf("Resume = %v, want %v", got.Resume, tt.expected.Resume)
			}
			if got.Verbose != tt.expected.Verbose {
				t.Errorf("Verbose = %v, want %v", got.Verbose, tt.expected.Verbose)
			}
			if got.Help != tt.expected.Help {
				t.Errorf("Help = %v, want %v", got.Help, tt.expected.Help)
			}
			if got.Web != tt.expected.Web {
				t.Errorf("Web = %v, want %v", got.Web, tt.expected.Web)
			}
			if got.WebPort != tt.expected.WebPort {
				t.Errorf("WebPort = %v, want %v", got.WebPort, tt.expected.WebPort)
			}
			if got.SkipCleanup != tt.expected.SkipCleanup {
				t.Errorf("SkipCleanup = %v, want %v", got.SkipCleanup, tt.expected.SkipCleanup)
			}
			if got.AutoApprove != tt.expected.AutoApprove {
				t.Errorf("AutoApprove = %v, want %v", got.AutoApprove, tt.expected.AutoApprove)
			}
			if len(got.UnknownFlags) != len(tt.expected.UnknownFlags) {
				t.Errorf("UnknownFlags length = %d, want %d", len(got.UnknownFlags), len(tt.expected.UnknownFlags))
			}
			for i, flag := range got.UnknownFlags {
				if flag != tt.expected.UnknownFlags[i] {
					t.Errorf("UnknownFlags[%d] = %q, want %q", i, flag, tt.expected.UnknownFlags[i])
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{name: "help bypasses validation", opts: Options{Help: true}, wantErr: false},
		{name: "status bypasses validation", opts: Options{Status: true}, wantErr: false},
		{name: "status rejects yolo", opts: Options{Status: true, AutoApprove: true}, wantErr: true},
		{name: "clean bypasses validation", opts: Options{Clean: true}, wantErr: false},
		{name: "clean rejects yolo", opts: Options{Clean: true, AutoApprove: true}, wantErr: true},
		{name: "clean with unknown flags bypasses validation", opts: Options{Clean: true, UnknownFlags: []string{"--bogus"}}, wantErr: false},
		{name: "version bypasses validation", opts: Options{Version: true}, wantErr: false},
		{name: "version rejects yolo", opts: Options{Version: true, AutoApprove: true}, wantErr: true},
		{name: "version with unknown flags bypasses validation", opts: Options{Version: true, UnknownFlags: []string{"--bogus"}}, wantErr: false},
		{name: "update bypasses validation", opts: Options{Update: true}, wantErr: false},
		{name: "update rejects yolo", opts: Options{Update: true, AutoApprove: true}, wantErr: true},
		{name: "dry run rejects yolo", opts: Options{DryRun: true, AutoApprove: true}, wantErr: true},
		{name: "web rejects yolo", opts: Options{Web: true, AutoApprove: true}, wantErr: true},
		{name: "resume without prompt is valid", opts: Options{Resume: true}, wantErr: false},
		{name: "resume with yolo is valid", opts: Options{Resume: true, AutoApprove: true}, wantErr: false},
		{name: "prompt provided is valid", opts: Options{Prompt: "do something"}, wantErr: false},
		{name: "no prompt no resume is valid", opts: Options{}, wantErr: false},
		{name: "unknown flags invalid without subcommand", opts: Options{UnknownFlags: []string{"--bogus"}}, wantErr: true},
		{name: "empty prompt with dry run is valid", opts: Options{DryRun: true}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHelpText(t *testing.T) {
	text := HelpText()
	for _, phrase := range []string{"Ralph", "Usage:", "Options:", "Environment:", "RALPH_RUNNER", "default: claude", "--dry-run", "--resume", "--verbose", "-v", "--help", "status", "ralph clean", "ralph version", "ralph update", "--ref", "--check", "RALPH_REPO", "ralph web", "--port", "8080", "# TUI prompt screen (requires a terminal)", "ralph --dry-run", "--skip-cleanup", "--yolo"} {
		if !strings.Contains(text, phrase) {
			t.Errorf("HelpText() missing %q", phrase)
		}
	}
}
