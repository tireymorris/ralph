package story

import (
	"context"
	"errors"
	"testing"

	"ralph/internal/config"
	"ralph/internal/prd"
	"ralph/internal/runner"
)

func TestNewImplementer(t *testing.T) {
	cfg := config.DefaultConfig()
	impl := NewImplementer(cfg)

	if impl == nil {
		t.Fatal("NewImplementer() returned nil")
	}
	if impl.cfg != cfg {
		t.Error("cfg not set correctly")
	}
	if impl.runner == nil {
		t.Error("runner should not be nil")
	}
	if impl.git == nil {
		t.Error("git should not be nil")
	}
}

type mockRunner struct {
	result *runner.Result
	err    error
}

func (m *mockRunner) RunOpenCode(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) (*runner.Result, error) {
	return m.result, m.err
}

type mockGit struct {
	err error
}

func (m *mockGit) CommitStory(storyID, title, description string) error {
	return m.err
}

func TestNewImplementerWithDeps(t *testing.T) {
	cfg := config.DefaultConfig()
	r := &mockRunner{}
	g := &mockGit{}

	impl := NewImplementerWithDeps(cfg, r, g)

	if impl == nil {
		t.Fatal("NewImplementerWithDeps() returned nil")
	}
	if impl.cfg != cfg {
		t.Error("cfg not set correctly")
	}
}

func TestImplementRunnerError(t *testing.T) {
	cfg := config.DefaultConfig()
	r := &mockRunner{err: errors.New("runner error")}
	g := &mockGit{}

	impl := NewImplementerWithDeps(cfg, r, g)

	story := &prd.Story{
		ID:                 "s1",
		Title:              "Test",
		Description:        "Desc",
		AcceptanceCriteria: []string{"ac"},
		TestSpec:           "spec",
	}
	p := &prd.PRD{Stories: []*prd.Story{story}}

	success, err := impl.Implement(context.Background(), story, 1, p, nil)
	if err == nil {
		t.Error("Implement() should return error")
	}
	if success {
		t.Error("success should be false on error")
	}
}

func TestImplementResultError(t *testing.T) {
	cfg := config.DefaultConfig()
	r := &mockRunner{result: &runner.Result{Error: errors.New("result error")}}
	g := &mockGit{}

	impl := NewImplementerWithDeps(cfg, r, g)

	story := &prd.Story{
		ID:                 "s1",
		Title:              "Test",
		Description:        "Desc",
		AcceptanceCriteria: []string{"ac"},
	}
	p := &prd.PRD{Stories: []*prd.Story{story}}

	success, err := impl.Implement(context.Background(), story, 1, p, nil)
	if err != nil {
		t.Errorf("Implement() error = %v", err)
	}
	if success {
		t.Error("success should be false on result error")
	}
}

func TestImplementNoCompleted(t *testing.T) {
	cfg := config.DefaultConfig()
	r := &mockRunner{result: &runner.Result{Output: "no completed marker"}}
	g := &mockGit{}

	impl := NewImplementerWithDeps(cfg, r, g)

	story := &prd.Story{
		ID:                 "s1",
		Title:              "Test",
		Description:        "Desc",
		AcceptanceCriteria: []string{"ac"},
	}
	p := &prd.PRD{Stories: []*prd.Story{story}}

	success, err := impl.Implement(context.Background(), story, 1, p, nil)
	if err != nil {
		t.Errorf("Implement() error = %v", err)
	}
	if success {
		t.Error("success should be false without COMPLETED marker")
	}
}

func TestImplementSuccess(t *testing.T) {
	cfg := config.DefaultConfig()
	r := &mockRunner{result: &runner.Result{Output: "COMPLETED: done"}}
	g := &mockGit{}

	impl := NewImplementerWithDeps(cfg, r, g)

	story := &prd.Story{
		ID:                 "s1",
		Title:              "Test",
		Description:        "Desc",
		AcceptanceCriteria: []string{"ac"},
	}
	p := &prd.PRD{Stories: []*prd.Story{story}}

	success, err := impl.Implement(context.Background(), story, 1, p, nil)
	if err != nil {
		t.Errorf("Implement() error = %v", err)
	}
	if !success {
		t.Error("success should be true")
	}
}

func TestImplementCommitError(t *testing.T) {
	cfg := config.DefaultConfig()
	r := &mockRunner{result: &runner.Result{Output: "COMPLETED: done"}}
	g := &mockGit{err: errors.New("commit error")}

	impl := NewImplementerWithDeps(cfg, r, g)

	story := &prd.Story{
		ID:                 "s1",
		Title:              "Test",
		Description:        "Desc",
		AcceptanceCriteria: []string{"ac"},
	}
	p := &prd.PRD{Stories: []*prd.Story{story}}

	outputCh := make(chan runner.OutputLine, 10)
	success, err := impl.Implement(context.Background(), story, 1, p, outputCh)
	if err != nil {
		t.Errorf("Implement() should not return error, got %v", err)
	}
	if !success {
		t.Error("success should still be true even with commit error")
	}
}

func TestImplementWithOutputChannel(t *testing.T) {
	cfg := config.DefaultConfig()
	r := &mockRunner{result: &runner.Result{Output: "COMPLETED: done"}}
	g := &mockGit{}

	impl := NewImplementerWithDeps(cfg, r, g)

	story := &prd.Story{
		ID:                 "s1",
		Title:              "Test",
		Description:        "Desc",
		AcceptanceCriteria: []string{"ac"},
	}
	p := &prd.PRD{Stories: []*prd.Story{story}}

	outputCh := make(chan runner.OutputLine, 10)
	success, _ := impl.Implement(context.Background(), story, 1, p, outputCh)
	if !success {
		t.Error("success should be true")
	}
}

func TestIsCompletionMarkerPresent(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{
			name:   "valid completion at start of line",
			output: "COMPLETED: done",
			want:   true,
		},
		{
			name:   "valid completion with leading whitespace",
			output: "  COMPLETED: summary here",
			want:   true,
		},
		{
			name:   "valid completion in multi-line output",
			output: "some output\nmore output\nCOMPLETED: all done\n",
			want:   true,
		},
		{
			name:   "quoted completion marker",
			output: `"COMPLETED: done"`,
			want:   true,
		},
		{
			name:   "no completion marker",
			output: "no marker here",
			want:   false,
		},
		{
			name:   "NOT COMPLETED should not match",
			output: "NOT COMPLETED: failed",
			want:   false,
		},
		{
			name:   "UNCOMPLETED should not match",
			output: "UNCOMPLETED: something",
			want:   false,
		},
		{
			name:   "completion in middle of word should not match",
			output: "xCOMPLETED: no",
			want:   false,
		},
		{
			name:   "empty output",
			output: "",
			want:   false,
		},
		{
			name:   "completion marker only",
			output: "COMPLETED:",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCompletionMarkerPresent(tt.output)
			if got != tt.want {
				t.Errorf("isCompletionMarkerPresent(%q) = %v, want %v", tt.output, got, tt.want)
			}
		})
	}
}
