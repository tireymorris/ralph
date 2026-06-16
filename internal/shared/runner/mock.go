package runner

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"time"

	promptpkg "ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
)

const mockQuestionsFile = ".ralph/questions.json"

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
	switch promptpkg.Kind(prompt) {
	case promptpkg.KindClarify:
		questions := "[]"
		if env := os.Getenv("RALPH_MOCK_QUESTIONS"); env != "" {
			questions = env
		}
		return writeStateFile(filepath.Join(workDir, mockQuestionsFile), questions)

	case promptpkg.KindPRDSelfReview:
		verdict := `{"approved":true,"summary":"mock self-review approved"}`
		return writeStateFile(filepath.Join(workDir, promptpkg.PRDSelfReviewVerdictFile), verdict)

	case promptpkg.KindPRDGenerate, promptpkg.KindPRDCritiqueRevision, promptpkg.KindPRDClarificationRevision, promptpkg.KindFollowUp:
		prdPath := m.cfg.PRDPath()
		data := `{"version":1,"project_name":"Mock","branch_name":"main","stories":[{"id":"story-1","title":"Mock story","description":"d","slices":[{"id":"slice-1","behavior":"first behavior","red_hint":"write failing test for first behavior"},{"id":"slice-2","behavior":"second behavior","red_hint":"write failing test for second behavior","refactor_hint":"extract helper"}],"priority":1}]}`
		return os.WriteFile(prdPath, []byte(data), 0o644)

	case promptpkg.KindDiffReview:
		findings := "===ralph-findings===\n[]\n===/ralph-findings==="
		select {
		case outputCh <- OutputLine{Text: findings, Time: time.Now()}:
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil

	case promptpkg.KindStoryImplement:
		if d := os.Getenv("RALPH_MOCK_IMPL_DELAY_MS"); d != "" {
			if ms, err := strconv.Atoi(d); err == nil && ms > 0 {
				select {
				case <-time.After(time.Duration(ms) * time.Millisecond):
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
		return advanceMockSlice(m.cfg)
	}

	return nil
}

func advanceMockSlice(cfg *config.Config) error {
	p, err := prd.Load(cfg)
	if err != nil {
		return err
	}

	story := p.NextReadyStory()
	if story == nil {
		story = p.NextPendingStory()
	}
	if story == nil {
		return nil
	}
	slice := story.NextPendingSlice()
	if slice == nil {
		if !story.Passes {
			story.Passes = story.AllSlicesPassed()
		}
		return prd.Save(cfg, p)
	}

	slice.Passes = true
	story.Passes = story.AllSlicesPassed()
	return prd.Save(cfg, p)
}

// writeStateFile mimics a real agent creating parent directories before writing.
func writeStateFile(path, contents string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(contents), 0o644)
}
