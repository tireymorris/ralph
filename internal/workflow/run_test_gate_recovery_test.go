package workflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/prd/prdtest"
	"ralph/internal/shared/runner"
	"ralph/internal/shared/testgit"
)

func TestRunTestGateWithRecoveryFixesFailingTests(t *testing.T) {
	workDir, _ := testgit.RepoWithWorkingTreeDiff(t)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"
	cfg.AutoApprove = true
	cfg.TestCommand = "test -f pkg/greet/greet.go"
	cfg.SkipCleanup = true

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories: []*prd.Story{
			{ID: "story-1", Title: "One", Description: "d", Slices: prdtest.Slices("a"), Priority: 1, Passes: false},
		},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("save PRD: %v", err)
	}

	greetDir := filepath.Join(workDir, "pkg", "greet")
	if err := os.MkdirAll(greetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(greetDir, "greet_test.go"), []byte(`package greet

import "testing"

func TestHello(t *testing.T) {
	if got := Hello(); got != "hello" {
		t.Fatalf("Hello() = %q, want %q", got, "hello")
	}
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(_ context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if isRecoveryPrompt(p) {
			return os.WriteFile(filepath.Join(greetDir, "greet.go"), []byte(`package greet

func Hello() string { return "hello" }
`), 0o644)
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	if err := exec.runTestGateWithRecovery(context.Background(), testPRD); err != nil {
		t.Fatalf("runTestGateWithRecovery() error = %v", err)
	}
}
