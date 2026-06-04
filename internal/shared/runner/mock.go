package runner

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ralph/internal/shared/config"
)

const mockQuestionsFile = ".ralph_questions.json"

// Mock is a deterministic runner for integration tests (RALPH_RUNNER=mock).
type Mock struct {
	cfg *config.Config
}

func NewMock(cfg *config.Config) RunnerInterface {
	return &Mock{cfg: cfg}
}

func (m *Mock) RunnerName() string  { return "mock" }
func (m *Mock) CommandName() string { return "mock" }
func (m *Mock) IsInternalLog(string) bool {
	return false
}

func (m *Mock) Run(ctx context.Context, prompt string, outputCh chan<- OutputLine) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	line := OutputLine{Text: "mock runner output", Time: time.Now()}
	select {
	case outputCh <- line:
	case <-ctx.Done():
		return ctx.Err()
	}

	workDir := m.cfg.WorkDir
	if strings.Contains(prompt, mockQuestionsFile) {
		questions := "[]"
		if env := os.Getenv("RALPH_MOCK_QUESTIONS"); env != "" {
			questions = env
		}
		path := filepath.Join(workDir, mockQuestionsFile)
		return os.WriteFile(path, []byte(questions), 0o644)
	}

	if strings.Contains(prompt, "Write the PRD file") || strings.Contains(prompt, "Write the updated PRD file") {
		prdPath := m.cfg.PRDPath()
		data := `{"version":1,"project_name":"Mock","branch_name":"main","stories":[{"id":"story-1","title":"Mock story","description":"d","acceptance_criteria":["ok"],"priority":1}]}`
		return os.WriteFile(prdPath, []byte(data), 0o644)
	}

	if strings.Contains(prompt, "===ralph-findings===") {
		findings := "===ralph-findings===\n[]\n===/ralph-findings==="
		select {
		case outputCh <- OutputLine{Text: findings, Time: time.Now()}:
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	}

	if d := os.Getenv("RALPH_MOCK_IMPL_DELAY_MS"); d != "" {
		if ms, err := strconv.Atoi(d); err == nil && ms > 0 {
			select {
			case <-time.After(time.Duration(ms) * time.Millisecond):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return nil
}
