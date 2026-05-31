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
		{name: "resume flag", args: []string{"--resume"}, expected: Options{Resume: true}},
		{name: "verbose flag short", args: []string{"-v"}, expected: Options{Verbose: true}},
		{name: "verbose flag long", args: []string{"--verbose"}, expected: Options{Verbose: true}},
		{name: "single prompt word", args: []string{"hello"}, expected: Options{Prompt: "hello"}},
		{name: "multi word prompt", args: []string{"hello", "world"}, expected: Options{Prompt: "hello world"}},
		{name: "prompt with flags", args: []string{"Add", "feature", "--dry-run"}, expected: Options{Prompt: "Add feature", DryRun: true}},
		{name: "unknown flag captured", args: []string{"--unknown", "prompt"}, expected: Options{Prompt: "prompt", UnknownFlags: []string{"--unknown"}}},
		{name: "multiple unknown flags captured", args: []string{"--foo", "-x", "prompt", "--bar"}, expected: Options{Prompt: "prompt", UnknownFlags: []string{"--foo", "-x", "--bar"}}},
		{name: "status command", args: []string{"status"}, expected: Options{Status: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.args)
			if got.Status != tt.expected.Status {
				t.Errorf("Status = %v, want %v", got.Status, tt.expected.Status)
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
		{name: "resume without prompt is valid", opts: Options{Resume: true}, wantErr: false},
		{name: "prompt provided is valid", opts: Options{Prompt: "do something"}, wantErr: false},
		{name: "no prompt no resume is valid", opts: Options{}, wantErr: false},
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
	for _, phrase := range []string{"Ralph", "Usage:", "Options:", "Environment:", "RALPH_RUNNER", "default: claude", "--dry-run", "--resume", "--verbose", "-v", "--help", "status", "# TUI prompt screen (requires a terminal)", "ralph --dry-run"} {
		if !strings.Contains(text, phrase) {
			t.Errorf("HelpText() missing %q", phrase)
		}
	}
}
