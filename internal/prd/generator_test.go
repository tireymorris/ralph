package prd

import (
	"context"
	"fmt"
	"testing"

	"ralph/internal/config"
	"ralph/internal/runner"
)

func TestNewGenerator(t *testing.T) {
	cfg := config.DefaultConfig()
	gen := NewGenerator(cfg)

	if gen == nil {
		t.Fatal("NewGenerator() returned nil")
	}
	if gen.cfg != cfg {
		t.Error("cfg not set correctly")
	}
	if gen.runner == nil {
		t.Error("runner should not be nil")
	}
}

type mockRunner struct {
	result *runner.Result
	err    error
}

func (m *mockRunner) RunOpenCode(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) (*runner.Result, error) {
	return m.result, m.err
}

func TestNewGeneratorWithRunner(t *testing.T) {
	cfg := config.DefaultConfig()
	r := &mockRunner{}
	gen := NewGeneratorWithRunner(cfg, r)

	if gen == nil {
		t.Fatal("NewGeneratorWithRunner() returned nil")
	}
}

func TestGenerateRunnerError(t *testing.T) {
	cfg := config.DefaultConfig()
	r := &mockRunner{err: fmt.Errorf("runner error")}
	gen := NewGeneratorWithRunner(cfg, r)

	_, err := gen.Generate(context.Background(), "test", nil)
	if err == nil {
		t.Error("Generate() should return error")
	}
}

func TestGenerateResultError(t *testing.T) {
	cfg := config.DefaultConfig()
	r := &mockRunner{result: &runner.Result{Error: fmt.Errorf("result error")}}
	gen := NewGeneratorWithRunner(cfg, r)

	_, err := gen.Generate(context.Background(), "test", nil)
	if err == nil {
		t.Error("Generate() should return error")
	}
}

func TestGenerateParseError(t *testing.T) {
	cfg := config.DefaultConfig()
	r := &mockRunner{result: &runner.Result{Output: "not valid json"}}
	gen := NewGeneratorWithRunner(cfg, r)

	_, err := gen.Generate(context.Background(), "test", nil)
	if err == nil {
		t.Error("Generate() should return error on parse failure")
	}
}

func TestGenerateValidationError(t *testing.T) {
	cfg := config.DefaultConfig()
	r := &mockRunner{result: &runner.Result{Output: `{"stories":[]}`}}
	gen := NewGeneratorWithRunner(cfg, r)

	_, err := gen.Generate(context.Background(), "test", nil)
	if err == nil {
		t.Error("Generate() should return error on validation failure")
	}
}

func TestGenerateSuccess(t *testing.T) {
	cfg := config.DefaultConfig()
	validPRD := `{
		"project_name": "Test",
		"stories": [
			{"id": "2", "title": "T2", "description": "D2", "acceptance_criteria": ["a"], "priority": 2},
			{"id": "1", "title": "T1", "description": "D1", "acceptance_criteria": ["a"], "priority": 1}
		]
	}`
	r := &mockRunner{result: &runner.Result{Output: validPRD}}
	gen := NewGeneratorWithRunner(cfg, r)

	p, err := gen.Generate(context.Background(), "test", nil)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if p.ProjectName != "Test" {
		t.Errorf("ProjectName = %q, want %q", p.ProjectName, "Test")
	}
	if p.Stories[0].Priority != 1 {
		t.Error("Stories should be sorted by priority")
	}
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name        string
		response    string
		wantProject string
		wantErr     bool
	}{
		{
			name:        "valid json",
			response:    `{"project_name": "Test", "stories": [{"id": "1", "title": "T", "description": "D", "acceptance_criteria": ["a"], "priority": 1}]}`,
			wantProject: "Test",
			wantErr:     false,
		},
		{
			name:        "json with surrounding text",
			response:    `Here is the PRD: {"project_name": "Test", "stories": [{"id": "1", "title": "T", "description": "D", "acceptance_criteria": ["a"], "priority": 1}]} That's all.`,
			wantProject: "Test",
			wantErr:     false,
		},
		{
			name:        "json with leading whitespace",
			response:    `   {"project_name": "Test", "stories": []}`,
			wantProject: "Test",
			wantErr:     false,
		},
		{
			name:        "json with braces in string values",
			response:    `{"project_name": "Test {Project}", "stories": [{"id": "1", "title": "Add func() { }", "description": "Implement { and }", "acceptance_criteria": ["code has {}"], "priority": 1}]}`,
			wantProject: "Test {Project}",
			wantErr:     false,
		},
		{
			name:        "json with escaped quotes",
			response:    `{"project_name": "Test \"Quoted\"", "stories": [{"id": "1", "title": "T", "description": "D", "acceptance_criteria": ["a"], "priority": 1}]}`,
			wantProject: `Test "Quoted"`,
			wantErr:     false,
		},
		{
			name:     "no json object",
			response: "no json here",
			wantErr:  true,
		},
		{
			name:     "incomplete json",
			response: `{"project_name": "Test"`,
			wantErr:  true,
		},
		{
			name:     "invalid json syntax",
			response: `{"project_name": Test}`,
			wantErr:  true,
		},
		{
			name:     "empty response",
			response: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseResponse(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got.ProjectName != tt.wantProject {
				t.Errorf("parseResponse() project = %q, want %q", got.ProjectName, tt.wantProject)
			}
		})
	}
}

func TestFindMatchingBrace(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		start int
		want  int
	}{
		{
			name:  "simple object",
			s:     "{}",
			start: 0,
			want:  2,
		},
		{
			name:  "nested object",
			s:     `{"a": {"b": 1}}`,
			start: 0,
			want:  15,
		},
		{
			name:  "with array",
			s:     `{"a": [1, 2]}`,
			start: 0,
			want:  13,
		},
		{
			name:  "multiple objects finds first",
			s:     `{"a": 1}{"b": 2}`,
			start: 0,
			want:  8,
		},
		{
			name:  "unmatched brace",
			s:     `{"a": 1`,
			start: 0,
			want:  -1,
		},
		{
			name:  "start from middle",
			s:     `xxx{"a": 1}yyy`,
			start: 3,
			want:  11,
		},
		{
			name:  "deeply nested",
			s:     `{"a": {"b": {"c": {"d": 1}}}}`,
			start: 0,
			want:  29,
		},
		{
			name:  "brace inside string",
			s:     `{"a": "value with { and } braces"}`,
			start: 0,
			want:  34,
		},
		{
			name:  "nested braces in string",
			s:     `{"code": "func() { return {}; }"}`,
			start: 0,
			want:  33,
		},
		{
			name:  "escaped quote in string",
			s:     `{"a": "he said \"hello\""}`,
			start: 0,
			want:  26,
		},
		{
			name:  "escaped backslash before quote",
			s:     `{"a": "path\\"}`,
			start: 0,
			want:  15,
		},
		{
			name:  "mixed escapes and braces",
			s:     `{"a": "test \" { } \\"}`,
			start: 0,
			want:  23,
		},
		{
			name:  "multiple strings with braces",
			s:     `{"a": "{", "b": "}"}`,
			start: 0,
			want:  20,
		},
		{
			name:  "complex nested with strings",
			s:     `{"obj": {"inner": "value with }"}, "other": 1}`,
			start: 0,
			want:  46,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findMatchingBrace(tt.s, tt.start)
			if got != tt.want {
				t.Errorf("findMatchingBrace() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		prd     *PRD
		wantErr bool
	}{
		{
			name: "valid prd",
			prd: &PRD{
				ProjectName: "Test",
				Stories: []*Story{
					{
						ID:                 "1",
						Title:              "Title",
						Description:        "Desc",
						AcceptanceCriteria: []string{"a"},
						Priority:           1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing project name",
			prd: &PRD{
				Stories: []*Story{
					{ID: "1", Title: "T", Description: "D", AcceptanceCriteria: []string{"a"}, Priority: 1},
				},
			},
			wantErr: true,
		},
		{
			name:    "no stories",
			prd:     &PRD{ProjectName: "Test", Stories: []*Story{}},
			wantErr: true,
		},
		{
			name: "story missing id",
			prd: &PRD{
				ProjectName: "Test",
				Stories:     []*Story{{Title: "T", Description: "D", AcceptanceCriteria: []string{"a"}, Priority: 1}},
			},
			wantErr: true,
		},
		{
			name: "story missing title",
			prd: &PRD{
				ProjectName: "Test",
				Stories:     []*Story{{ID: "1", Description: "D", AcceptanceCriteria: []string{"a"}, Priority: 1}},
			},
			wantErr: true,
		},
		{
			name: "story missing description",
			prd: &PRD{
				ProjectName: "Test",
				Stories:     []*Story{{ID: "1", Title: "T", AcceptanceCriteria: []string{"a"}, Priority: 1}},
			},
			wantErr: true,
		},
		{
			name: "story missing acceptance criteria",
			prd: &PRD{
				ProjectName: "Test",
				Stories:     []*Story{{ID: "1", Title: "T", Description: "D", Priority: 1}},
			},
			wantErr: true,
		},
		{
			name: "story empty acceptance criteria",
			prd: &PRD{
				ProjectName: "Test",
				Stories:     []*Story{{ID: "1", Title: "T", Description: "D", AcceptanceCriteria: []string{}, Priority: 1}},
			},
			wantErr: true,
		},
		{
			name: "story missing priority",
			prd: &PRD{
				ProjectName: "Test",
				Stories:     []*Story{{ID: "1", Title: "T", Description: "D", AcceptanceCriteria: []string{"a"}}},
			},
			wantErr: true,
		},
		{
			name: "multiple stories second invalid",
			prd: &PRD{
				ProjectName: "Test",
				Stories: []*Story{
					{ID: "1", Title: "T", Description: "D", AcceptanceCriteria: []string{"a"}, Priority: 1},
					{ID: "2", Description: "D", AcceptanceCriteria: []string{"a"}, Priority: 2},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.prd)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
