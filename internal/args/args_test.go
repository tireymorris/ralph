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
		{name: "empty args", args: []string{}, expected: Options{Mode: ModeAuto}},
		{name: "help flag short", args: []string{"-h"}, expected: Options{Help: true, Mode: ModeAuto}},
		{name: "help flag long", args: []string{"--help"}, expected: Options{Help: true, Mode: ModeAuto}},
		{name: "dry run flag", args: []string{"--dry-run"}, expected: Options{DryRun: true, Mode: ModeAuto}},
		{name: "resume flag", args: []string{"--resume"}, expected: Options{Resume: true, Mode: ModeAuto}},
		{name: "verbose flag short", args: []string{"-v"}, expected: Options{Verbose: true, Mode: ModeAuto}},
		{name: "verbose flag long", args: []string{"--verbose"}, expected: Options{Verbose: true, Mode: ModeAuto}},
		{name: "single prompt word", args: []string{"hello"}, expected: Options{Prompt: "hello", Mode: ModeAuto}},
		{name: "multi word prompt", args: []string{"hello", "world"}, expected: Options{Prompt: "hello world", Mode: ModeAuto}},
		{name: "prompt with flags", args: []string{"Add", "feature", "--dry-run"}, expected: Options{Prompt: "Add feature", DryRun: true, Mode: ModeAuto}},
		{name: "unknown flag captured", args: []string{"--unknown", "prompt"}, expected: Options{Prompt: "prompt", Mode: ModeAuto, UnknownFlags: []string{"--unknown"}}},
		{name: "multiple unknown flags captured", args: []string{"--foo", "-x", "prompt", "--bar"}, expected: Options{Prompt: "prompt", Mode: ModeAuto, UnknownFlags: []string{"--foo", "-x", "--bar"}}},
		{name: "status command", args: []string{"status"}, expected: Options{Mode: ModeStatus}},
		{name: "prd command", args: []string{"prd"}, expected: Options{Mode: ModePRD}},
		{name: "review command", args: []string{"review"}, expected: Options{Mode: ModeReview}},
		{name: "implement command", args: []string{"implement"}, expected: Options{Mode: ModeImplement}},
		{name: "run command", args: []string{"run"}, expected: Options{Mode: ModeAuto}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.args)
			if got.Mode != tt.expected.Mode {
				t.Errorf("Mode = %q, want %q", got.Mode, tt.expected.Mode)
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
		{name: "resume without prompt is valid", opts: Options{Resume: true}, wantErr: false},
		{name: "prompt provided is valid", opts: Options{Prompt: "do something"}, wantErr: false},
		{name: "no prompt no resume is invalid", opts: Options{}, wantErr: true},
		{name: "status command without prompt is valid", opts: Options{Mode: ModeStatus}, wantErr: false},
		{name: "review command without prompt is valid", opts: Options{Mode: ModeReview}, wantErr: false},
		{name: "implement command without prompt is valid", opts: Options{Mode: ModeImplement}, wantErr: false},
		{name: "prd command without prompt is invalid", opts: Options{Mode: ModePRD}, wantErr: true},
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
	for _, phrase := range []string{"Ralph", "Usage:", "Options:", "--dry-run", "--resume", "--verbose", "-v", "--help", "run", "status", "prd", "implement"} {
		if !strings.Contains(text, phrase) {
			t.Errorf("HelpText() missing %q", phrase)
		}
	}
}
