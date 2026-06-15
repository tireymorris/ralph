package app

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"ralph/internal/args"
	"ralph/internal/shared/config"
	"ralph/internal/update"
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
		wantRunHeadless    bool
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
			name:               "headless",
			opts:               &args.Options{Headless: true, Prompt: "build a feature", AutoApprove: true},
			wantCode:           9,
			wantLoadConfig:     true,
			wantValidateResume: true,
			wantValidateGit:    true,
			wantRunHeadless:    true,
			wantPrompt:         "build a feature",
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
				runHeadless []struct {
					prompt string
					resume bool
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
				runHeadless: func(*config.Config, string, bool) int {
					calls.runHeadless = append(calls.runHeadless, struct {
						prompt string
						resume bool
					}{prompt: tt.opts.Prompt, resume: tt.opts.Resume})
					return 9
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
			if got := len(calls.runHeadless) > 0; got != tt.wantRunHeadless {
				t.Fatalf("runHeadless called = %v, want %v", got, tt.wantRunHeadless)
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
			if tt.wantRunHeadless {
				got := calls.runHeadless[0]
				if got.prompt != tt.wantPrompt || got.resume != tt.wantResume {
					t.Fatalf("runHeadless args = %#v, want prompt=%q resume=%v", got, tt.wantPrompt, tt.wantResume)
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

func TestCoordinatorBootUpdatePrecedesPromptFlow(t *testing.T) {
	oldCheck := updateCheck
	oldInstall := updateInstall
	defer func() {
		updateCheck = oldCheck
		updateInstall = oldInstall
	}()

	calls := make([]string, 0, 4)
	updateCheck = func(context.Context, string, string) (bool, string, string, error) {
		calls = append(calls, "update-check")
		return false, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", nil
	}
	updateInstall = func(context.Context, update.InstallOptions) error {
		calls = append(calls, "update-install")
		return nil
	}

	c := &Coordinator{
		loadConfig: func() (*config.Config, error) {
			calls = append(calls, "load-config")
			return &config.Config{WorkDir: t.TempDir(), PRDFile: "prd.json"}, nil
		},
		runTUI: func(*config.Config, string, bool, bool, bool) int {
			calls = append(calls, "run-tui")
			return 0
		},
		validateGit: func(string) error {
			return nil
		},
		validateResume: func(*config.Config, bool) error {
			return nil
		},
		helpText:    func() string { return "help text" },
		versionInfo: func() string { return "version text" },
		isTerminal:  func(uintptr) bool { return true },
	}

	code, _, stderr := captureCoordinatorRun(t, c, &args.Options{Prompt: "build a feature"})
	if code != 0 {
		t.Fatalf("Run() = %d, want 0", code)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if len(calls) < 3 {
		t.Fatalf("calls = %v, want update and TUI flow", calls)
	}
	if calls[0] != "update-check" || calls[1] != "update-install" || calls[len(calls)-1] != "run-tui" {
		t.Fatalf("calls = %v, want boot update before prompt flow", calls)
	}
}

func TestCoordinatorBootUpdateFailureDoesNotBlockPromptFlow(t *testing.T) {
	oldCheck := updateCheck
	oldInstall := updateInstall
	defer func() {
		updateCheck = oldCheck
		updateInstall = oldInstall
	}()

	updateCheck = func(context.Context, string, string) (bool, string, string, error) {
		return false, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", nil
	}
	updateInstall = func(context.Context, update.InstallOptions) error {
		return context.Canceled
	}

	runTUICalled := false
	c := &Coordinator{
		loadConfig: func() (*config.Config, error) {
			return &config.Config{WorkDir: t.TempDir(), PRDFile: "prd.json"}, nil
		},
		runTUI: func(*config.Config, string, bool, bool, bool) int {
			runTUICalled = true
			return 0
		},
		validateGit: func(string) error {
			return nil
		},
		validateResume: func(*config.Config, bool) error {
			return nil
		},
		helpText:    func() string { return "help text" },
		versionInfo: func() string { return "version text" },
		isTerminal:  func(uintptr) bool { return true },
	}

	code, _, stderr := captureCoordinatorRun(t, c, &args.Options{Prompt: "build a feature"})
	if code != 0 {
		t.Fatalf("Run() = %d, want 0", code)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if !runTUICalled {
		t.Fatal("runTUI was not called after boot update failure")
	}
}

func TestCoordinatorExplicitUpdateSkipsBootUpdate(t *testing.T) {
	oldCheck := updateCheck
	oldInstall := updateInstall
	defer func() {
		updateCheck = oldCheck
		updateInstall = oldInstall
	}()

	calls := make([]string, 0, 2)
	updateCheck = func(context.Context, string, string) (bool, string, string, error) {
		calls = append(calls, "boot-check")
		return false, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", nil
	}
	updateInstall = func(context.Context, update.InstallOptions) error {
		calls = append(calls, "boot-install")
		return nil
	}

	c := &Coordinator{
		runUpdate: func(*args.Options) int {
			calls = append(calls, "run-update")
			return 0
		},
		loadConfig: func() (*config.Config, error) {
			t.Fatal("loadConfig should not run for explicit update")
			return nil, nil
		},
		helpText:    func() string { return "help text" },
		versionInfo: func() string { return "version text" },
		isTerminal:  func(uintptr) bool { return true },
	}

	code, stdout, stderr := captureCoordinatorRun(t, c, &args.Options{Update: true})
	if code != 0 {
		t.Fatalf("Run() = %d, want 0", code)
	}
	if stdout != "" || stderr != "" {
		t.Fatalf("stdout/stderr = %q / %q, want both empty", stdout, stderr)
	}
	if len(calls) != 1 || calls[0] != "run-update" {
		t.Fatalf("calls = %v, want only run-update", calls)
	}
}

func TestCoordinatorMetaCommandsSkipBootUpdate(t *testing.T) {
	oldCheck := updateCheck
	oldInstall := updateInstall
	defer func() {
		updateCheck = oldCheck
		updateInstall = oldInstall
	}()

	tests := []struct {
		name           string
		opts           *args.Options
		wantCode       int
		wantStdout     string
		wantRoute      bool
		wantRouteName  string
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
			name:          "update check",
			opts:          &args.Options{Update: true, UpdateCheck: true},
			wantCode:      8,
			wantRoute:     true,
			wantRouteName: "run-update-check",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := make([]string, 0, 2)
			updateCheck = func(context.Context, string, string) (bool, string, string, error) {
				calls = append(calls, "boot-check")
				return false, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", nil
			}
			updateInstall = func(context.Context, update.InstallOptions) error {
				calls = append(calls, "boot-install")
				return nil
			}

			c := &Coordinator{
				runUpdateCheck: func(*args.Options) int {
					calls = append(calls, "run-update-check")
					return 8
				},
				loadConfig: func() (*config.Config, error) {
					t.Fatal("loadConfig should not run for meta commands")
					return nil, nil
				},
				helpText:    func() string { return "help text" },
				versionInfo: func() string { return "version text" },
				isTerminal:  func(uintptr) bool { return true },
			}

			code, stdout, stderr := captureCoordinatorRun(t, c, tt.opts)
			if code != tt.wantCode {
				t.Fatalf("Run() = %d, want %d", code, tt.wantCode)
			}
			if stdout != tt.wantStdout || stderr != "" {
				t.Fatalf("stdout/stderr = %q / %q, want %q / %q", stdout, stderr, tt.wantStdout, "")
			}
			if tt.wantRoute {
				if len(calls) != 1 || calls[0] != tt.wantRouteName {
					t.Fatalf("calls = %v, want only %s", calls, tt.wantRouteName)
				}
			} else if len(calls) != 0 {
				t.Fatalf("calls = %v, want no boot update calls", calls)
			}
		})
	}
}
