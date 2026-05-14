package run

import (
	"context"
	"errors"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/prompt"
)

func TestNewCmd(t *testing.T) {
	cfg := config.DefaultConfig()
	c := NewCmd(cfg, "test prompt", true, false, false)

	if c == nil {
		t.Fatal("NewCmd() returned nil")
	}
	if c.cfg != cfg {
		t.Error("cfg not set correctly")
	}
	if c.prompt != "test prompt" {
		t.Errorf("prompt = %q, want %q", c.prompt, "test prompt")
	}
	if !c.dryRun {
		t.Error("dryRun should be true")
	}
	if c.resume {
		t.Error("resume should be false")
	}
}

func TestNewCmdResume(t *testing.T) {
	cfg := config.DefaultConfig()
	c := NewCmd(cfg, "", false, true, false)

	if !c.resume {
		t.Error("resume should be true")
	}
	if c.dryRun {
		t.Error("dryRun should be false")
	}
}

func TestNewCmdVerbose(t *testing.T) {
	cfg := config.DefaultConfig()
	c := NewCmd(cfg, "test", false, false, true)

	if !c.verbose {
		t.Error("verbose should be true")
	}
}

func TestRunReturnsErrorOnImplementationFailure(t *testing.T) {
	cfg := config.DefaultConfig()
	c := NewCmd(cfg, "test prompt", false, true, false)
	c.executor = &fakeExecutor{
		loadPRD: &prd.PRD{ProjectName: "Test", Stories: []*prd.Story{{ID: "1", Title: "Story", Description: "Desc", AcceptanceCriteria: []string{"AC"}, Priority: 1}}},
		runErr:  errors.New("implementation failed"),
	}

	code := c.Run()

	if code != 1 {
		t.Fatalf("Run() exit code = %d, want 1", code)
	}
}

type fakeExecutor struct {
	loadPRD *prd.PRD
	runErr  error
	calls   []string
}

func (f *fakeExecutor) RunClarify(ctx context.Context, userPrompt string) ([]prompt.QuestionAnswer, error) {
	f.calls = append(f.calls, "clarify")
	return nil, nil
}

func (f *fakeExecutor) RunLoad(ctx context.Context) (*prd.PRD, error) {
	f.calls = append(f.calls, "load")
	return f.loadPRD, nil
}

func (f *fakeExecutor) RunGenerateWithAnswers(ctx context.Context, userPrompt string, qas []prompt.QuestionAnswer) (*prd.PRD, error) {
	f.calls = append(f.calls, "generate")
	return f.loadPRD, nil
}

func (f *fakeExecutor) RunImplementation(ctx context.Context, p *prd.PRD) error {
	f.calls = append(f.calls, "implement")
	return f.runErr
}
