package prd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/prompt"
)

func TestRunWithMockExecutorWritesPRD(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"

	mockExec := &mockWorkflowExecutor{
		workDir: tmpDir,
		prdFile: "prd.json",
		prd: &prd.PRD{
			Version:     1,
			ProjectName: "Test Project",
			Stories: []*prd.Story{
				{ID: "story-1", Title: "Test Story", Description: "A test story", AcceptanceCriteria: []string{"It works"}, Priority: 1},
			},
		},
	}

	cmd := NewCmd(cfg, "build a test app", false)
	cmd.executor = mockExec

	code := cmd.Run()

	if code != 0 {
		t.Fatalf("Run() = %d, want 0", code)
	}

	prdPath := filepath.Join(tmpDir, "prd.json")
	data, err := os.ReadFile(prdPath)
	if err != nil {
		t.Fatalf("failed to read prd.json: %v", err)
	}

	var generated prd.PRD
	if err := json.Unmarshal(data, &generated); err != nil {
		t.Fatalf("prd.json is not valid JSON: %v", err)
	}

	if generated.ProjectName != "Test Project" {
		t.Errorf("ProjectName = %q, want %q", generated.ProjectName, "Test Project")
	}

	if len(generated.Stories) != 1 {
		t.Fatalf("len(Stories) = %d, want 1", len(generated.Stories))
	}

	if !mockExec.clarifyCalled {
		t.Error("RunClarify was not called")
	}

	if !mockExec.generateCalled {
		t.Error("RunGenerateWithAnswers was not called")
	}

	if mockExec.implementCalled {
		t.Error("RunImplementation should not be called by prd subcommand")
	}
}

type mockWorkflowExecutor struct {
	workDir         string
	prdFile         string
	prd             *prd.PRD
	clarifyErr      error
	generateErr     error
	clarifyCalled   bool
	generateCalled  bool
	implementCalled bool
}

func (m *mockWorkflowExecutor) RunClarify(ctx context.Context, userPrompt string) ([]prompt.QuestionAnswer, error) {
	m.clarifyCalled = true
	return nil, m.clarifyErr
}

func (m *mockWorkflowExecutor) RunGenerateWithAnswers(ctx context.Context, userPrompt string, qas []prompt.QuestionAnswer) (*prd.PRD, error) {
	m.generateCalled = true
	if m.prd != nil && m.workDir != "" {
		cfg := &config.Config{WorkDir: m.workDir, PRDFile: m.prdFile}
		if err := prd.Save(cfg, m.prd); err != nil {
			return nil, err
		}
	}
	return m.prd, m.generateErr
}

func (m *mockWorkflowExecutor) RunImplementation(ctx context.Context, p *prd.PRD) error {
	m.implementCalled = true
	return nil
}
