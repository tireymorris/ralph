package tui

import (
	"strings"
	"testing"

	"ralph/internal/config"
	"ralph/internal/prd"
)

func TestViewQuitting(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.quitting = true

	view := m.View()
	if !strings.Contains(view, "Goodbye") {
		t.Errorf("View() when quitting should say goodbye, got %q", view)
	}
}

func TestViewPhaseInit(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test prompt", false, false, false)
	m.phase = PhaseInit
	m.width = 80
	m.height = 24

	view := m.View()
	if !strings.Contains(view, "RALPH") {
		t.Error("View() should contain RALPH header")
	}
}

func TestViewPhasePRDGeneration(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test prompt", false, false, false)
	m.phase = PhasePRDGeneration
	m.width = 80
	m.height = 24

	view := m.View()
	if !strings.Contains(view, "test prompt") || !strings.Contains(view, "Generating") {
		t.Error("View() during PRD generation should show prompt and generating message")
	}
}

func TestViewPhaseImplementation(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhaseImplementation
	m.prd = &prd.PRD{
		ProjectName: "Test Project",
		BranchName:  "feature/test",
		Stories: []*prd.Story{
			{ID: "1", Title: "Story One", Passes: true},
			{ID: "2", Title: "Story Two", Passes: false},
		},
	}
	m.currentStory = m.prd.Stories[1]
	m.width = 80
	m.height = 24

	view := m.View()
	if !strings.Contains(view, "Test Project") {
		t.Error("View() should contain project name")
	}
	if !strings.Contains(view, "feature/test") {
		t.Error("View() should contain branch name")
	}
	if !strings.Contains(view, "Story One") {
		t.Error("View() should contain story titles")
	}
}

func TestViewPhaseImplementationNilPRD(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhaseImplementation
	m.prd = nil
	m.width = 80
	m.height = 24

	view := m.View()
	if view == "" {
		t.Error("View() should not be empty even with nil PRD")
	}
}

func TestViewPhaseCompletedDryRun(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", true, false, false)
	m.phase = PhaseCompleted
	m.dryRun = true
	m.width = 80
	m.height = 24

	view := m.View()
	if !strings.Contains(view, "Dry run") {
		t.Error("View() should mention dry run")
	}
}

func TestViewPhaseCompletedWithPRD(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhaseCompleted
	m.prd = &prd.PRD{
		ProjectName: "Done Project",
		Stories:     []*prd.Story{{ID: "1", Passes: true}},
	}
	m.iteration = 5
	m.width = 80
	m.height = 24

	view := m.View()
	if !strings.Contains(view, "Done Project") {
		t.Error("View() should show project name")
	}
	if !strings.Contains(view, "completed") {
		t.Error("View() should mention completed")
	}
}

func TestViewPhaseFailed(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhaseFailed
	m.width = 80
	m.height = 24

	view := m.View()
	if !strings.Contains(view, "failed") || !strings.Contains(view, "Failed") {
		t.Error("View() should mention failure")
	}
}

func TestViewPhaseFailedWithError(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhaseFailed
	m.err = &testError{msg: "test error"}
	m.width = 80
	m.height = 24

	view := m.View()
	if !strings.Contains(view, "test error") {
		t.Error("View() should show error message")
	}
}

func TestViewPhaseFailedWithFailedStories(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhaseFailed
	m.prd = &prd.PRD{
		Stories: []*prd.Story{
			{ID: "1", Title: "Failed Story", Passes: false, RetryCount: 3},
		},
	}
	m.width = 80
	m.height = 24

	view := m.View()
	if !strings.Contains(view, "Failed Story") {
		t.Error("View() should list failed stories")
	}
	if !strings.Contains(view, "--resume") {
		t.Error("View() should suggest resume option")
	}
}

func TestRenderHeader(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	header := m.renderHeader()
	if !strings.Contains(header, "RALPH") {
		t.Error("renderHeader() should contain RALPH")
	}
}

func TestRenderPhase(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	phases := []Phase{PhaseInit, PhasePRDGeneration, PhaseImplementation, PhaseCompleted, PhaseFailed}
	for _, p := range phases {
		m.phase = p
		result := m.renderPhase()
		if result == "" {
			t.Errorf("renderPhase() empty for phase %v", p)
		}
	}
}

func TestRenderLogsEmpty(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.logs = []string{}
	m.width = 80

	logs := m.renderLogs()
	if !strings.Contains(logs, "Waiting") {
		t.Error("renderLogs() with no logs should show waiting message")
	}
}

func TestRenderLogsWithContent(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.logs = []string{"Log line 1", "Log line 2"}
	m.width = 80

	logs := m.renderLogs()
	if !strings.Contains(logs, "Log line 1") {
		t.Error("renderLogs() should contain log lines")
	}
}

func TestRenderLogsTruncated(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.width = 80

	for i := 0; i < 20; i++ {
		m.logs = append(m.logs, "Log line")
	}

	logs := m.renderLogs()
	if logs == "" {
		t.Error("renderLogs() should not be empty")
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
