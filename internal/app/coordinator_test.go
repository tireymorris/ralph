package app

import (
	"bytes"
	"io"
	"os"
	"testing"

	"ralph/internal/args"
	"ralph/internal/shared/config"
)

func TestCoordinatorRoutesCommands(t *testing.T) {
	tests := []struct {
		name               string
		opts               *args.Options
		wantCode           int
		wantStdout         string
		wantStderr         string
		wantLoadConfig     bool
		wantValidateResume bool
		wantValidateGit    bool
		wantRunClean       bool
		wantRunStatus      bool
		wantRunWeb         bool
		wantRunTUI         bool
		wantRunUpdate      bool
		wantRunUpdateCheck bool
		wantResume         bool
		wantPrompt         string
		wantDryRun         bool
		wantVerbose        bool
		wantWebPort        int
		wantTerminalCheck  bool
	}{
		{
			name:       "help",
			opts:       &args.Options{Help: true},
			wantCode:   0,
			wantStdout: "help text",
		},
		{
			name:       "version",
			opts:       &args.Options{Version: true},
			wantCode:   0,
			wantStdout: "version text\n",
		},
		{
			name:          "update",
			opts:          &args.Options{Update: true},
			wantCode:      7,
			wantRunUpdate: true,
		},
		{
			name:               "update check",
			opts:               &args.Options{Update: true, UpdateCheck: true},
			wantCode:           8,
			wantRunUpdateCheck: true,
		},
		{
			name:               "clean",
			opts:               &args.Options{Clean: true, Resume: true},
			wantCode:           5,
			wantLoadConfig:     true,
			wantRunClean:       true,
			wantPrompt:         "",
			wantResume:         false,
			wantValidateResume: false,
		},
		{
			name:               "status",
			opts:               &args.Options{Status: true},
			wantCode:           4,
			wantLoadConfig:     true,
			wantValidateResume: true,
			wantRunStatus:      true,
		},
		{
			name:               "web",
			opts:               &args.Options{Web: true, WebPort: 3333},
			wantCode:           6,
			wantLoadConfig:     true,
			wantValidateResume: true,
			wantRunWeb:         true,
			wantWebPort:        3333,
		},
		{
			name:               "resume",
			opts:               &args.Options{Resume: true, Prompt: "build a feature", DryRun: true, Verbose: true},
			wantCode:           3,
			wantLoadConfig:     true,
			wantValidateResume: true,
			wantValidateGit:    true,
			wantRunTUI:         true,
			wantResume:         true,
			wantPrompt:         "build a feature",
			wantDryRun:         true,
			wantVerbose:        true,
			wantTerminalCheck:  true,
		},
		{
			name:               "non-tty prompt rejection",
			opts:               &args.Options{},
			wantCode:           1,
			wantLoadConfig:     true,
			wantValidateResume: true,
			wantTerminalCheck:  true,
			wantStderr:         "interactive prompt requires a terminal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{WorkDir: t.TempDir(), PRDFile: "prd.json"}
			calls := struct {
				loadConfig     int
				validateResume []bool
				validateGit    []string
				runClean       int
				runStatus      int
				runWeb         []int
				runTUI         []struct {
					prompt  string
					dryRun  bool
					resume  bool
					verbose bool
				}
				runUpdate      int
				runUpdateCheck int
			}{}

			c := &Coordinator{
				loadConfig: func() (*config.Config, error) {
					calls.loadConfig++
					return cfg, nil
				},
				runClean: func(*config.Config) int {
					calls.runClean++
					return 5
				},
				runStatus: func(*config.Config) int {
					calls.runStatus++
					return 4
				},
				runWeb: func(*config.Config, int) int {
					calls.runWeb = append(calls.runWeb, tt.opts.WebPort)
					return 6
				},
				runTUI: func(*config.Config, string, bool, bool, bool) int {
					calls.runTUI = append(calls.runTUI, struct {
						prompt  string
						dryRun  bool
						resume  bool
						verbose bool
					}{prompt: tt.opts.Prompt, dryRun: tt.opts.DryRun, resume: tt.opts.Resume, verbose: tt.opts.Verbose})
					return 3
				},
				runUpdate: func(*args.Options) int {
					calls.runUpdate++
					return 7
				},
				runUpdateCheck: func(*args.Options) int {
					calls.runUpdateCheck++
					return 8
				},
				validateGit: func(path string) error {
					calls.validateGit = append(calls.validateGit, path)
					return nil
				},
				validateResume: func(*config.Config, bool) error {
					calls.validateResume = append(calls.validateResume, tt.opts.Resume)
					return nil
				},
				helpText:    func() string { return "help text" },
				versionInfo: func() string { return "version text" },
				isTerminal: func(uintptr) bool {
					return !tt.wantTerminalCheck
				},
			}

			if tt.name == "non-tty prompt rejection" {
				c.isTerminal = func(uintptr) bool { return false }
			}

			code, stdout, stderr := captureCoordinatorRun(t, c, tt.opts)

			if code != tt.wantCode {
				t.Fatalf("Run() = %d, want %d", code, tt.wantCode)
			}
			if stdout != tt.wantStdout {
				t.Fatalf("stdout = %q, want %q", stdout, tt.wantStdout)
			}
			if tt.wantStderr != "" && !bytes.Contains([]byte(stderr), []byte(tt.wantStderr)) {
				t.Fatalf("stderr = %q, want substring %q", stderr, tt.wantStderr)
			}
			if got := calls.loadConfig > 0; got != tt.wantLoadConfig {
				t.Fatalf("loadConfig called = %v, want %v", got, tt.wantLoadConfig)
			}
			if got := len(calls.validateResume) > 0; got != tt.wantValidateResume {
				t.Fatalf("validateResume called = %v, want %v", got, tt.wantValidateResume)
			}
			if got := len(calls.validateGit) > 0; got != tt.wantValidateGit {
				t.Fatalf("validateGit called = %v, want %v", got, tt.wantValidateGit)
			}
			if got := calls.runClean > 0; got != tt.wantRunClean {
				t.Fatalf("runClean called = %v, want %v", got, tt.wantRunClean)
			}
			if got := calls.runStatus > 0; got != tt.wantRunStatus {
				t.Fatalf("runStatus called = %v, want %v", got, tt.wantRunStatus)
			}
			if got := len(calls.runWeb) > 0; got != tt.wantRunWeb {
				t.Fatalf("runWeb called = %v, want %v", got, tt.wantRunWeb)
			}
			if got := len(calls.runTUI) > 0; got != tt.wantRunTUI {
				t.Fatalf("runTUI called = %v, want %v", got, tt.wantRunTUI)
			}
			if got := calls.runUpdate > 0; got != tt.wantRunUpdate {
				t.Fatalf("runUpdate called = %v, want %v", got, tt.wantRunUpdate)
			}
			if got := calls.runUpdateCheck > 0; got != tt.wantRunUpdateCheck {
				t.Fatalf("runUpdateCheck called = %v, want %v", got, tt.wantRunUpdateCheck)
			}
			if tt.wantRunTUI {
				got := calls.runTUI[0]
				if got.prompt != tt.wantPrompt || got.dryRun != tt.wantDryRun || got.resume != tt.wantResume || got.verbose != tt.wantVerbose {
					t.Fatalf("runTUI args = %#v, want prompt=%q dryRun=%v resume=%v verbose=%v", got, tt.wantPrompt, tt.wantDryRun, tt.wantResume, tt.wantVerbose)
				}
			}
			if tt.wantRunWeb && calls.runWeb[0] != tt.wantWebPort {
				t.Fatalf("runWeb port = %d, want %d", calls.runWeb[0], tt.wantWebPort)
			}
		})
	}
}

func captureCoordinatorRun(t *testing.T, c *Coordinator, opts *args.Options) (int, string, string) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = wOut
	os.Stderr = wErr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	codeCh := make(chan int, 1)
	go func() {
		codeCh <- c.Run(opts)
		_ = wOut.Close()
		_ = wErr.Close()
	}()

	var stdoutBuf, stderrBuf bytes.Buffer
	if _, err := io.Copy(&stdoutBuf, rOut); err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(&stderrBuf, rErr); err != nil {
		t.Fatal(err)
	}
	code := <-codeCh
	return code, stdoutBuf.String(), stderrBuf.String()
}
