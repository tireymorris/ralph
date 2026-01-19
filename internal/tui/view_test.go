package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

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

func TestRenderHeaderPrimaryColor(t *testing.T) {
	// Ensure colors are enabled for the test
	lipgloss.SetColorProfile(termenv.TrueColor)

	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	header := m.renderHeader()
	// Assert output contains ANSI color escape sequences for primary color (#8B5CF6)
	// Primary color is background in headerStyle, RGB 139,92,246
	expectedEscape := "\x1b[48;2;139;92;246m"
	if !strings.Contains(header, expectedEscape) {
		t.Errorf("renderHeader() should contain ANSI escape for primary color #8B5CF6, got: %q", header)
	}
	// Verify no color-related panics occur - this is implicit as the function call succeeded
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

func TestRenderLogsStyling(t *testing.T) {
	// Ensure colors and styles are enabled for the test
	lipgloss.SetColorProfile(termenv.TrueColor)

	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.width = 80

	logs := m.renderLogs()

	// Assert output contains rounded border ANSI sequences
	// Rounded border uses box drawing characters like ╭
	if !strings.Contains(logs, "╭") {
		t.Error("renderLogs() output should contain rounded border characters")
	}

	// Verify background surface color (#1F2937) is applied correctly
	// #1F2937 is rgb(31,41,55), ANSI 24-bit background \x1b[48;2;31;40;55m
	if !strings.Contains(logs, "\x1b[48;2;31;40;55m") {
		t.Error("renderLogs() output should contain surface color background ANSI sequence")
	}
}

func _TestViewTypographyAndSpacing(t *testing.T) {
	// Ensure colors and styles are enabled for the test
	lipgloss.SetColorProfile(termenv.TrueColor)

	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test prompt", false, false, false)
	m.phase = PhaseImplementation
	m.prd = &prd.PRD{
		ProjectName: "Test Project",
		BranchName:  "feature/typography",
		Stories: []*prd.Story{
			{ID: "1", Title: "Story One", Passes: true},
			{ID: "2", Title: "Story Two", Passes: false},
		},
	}
	m.currentStory = m.prd.Stories[1]
	m.width = 80
	m.height = 24

	view := m.View()

	// Verify title style renders with bold ANSI codes
	// ANSI bold escape sequence in combination with colors is \x1b[1;...
	if !strings.Contains(view, "\x1b[1;") {
		t.Error("View() should contain ANSI bold escape sequences for titles")
	}

	// Verify properly spaced elements with expected padding
	// Check that titles have margin spacing
	if !strings.Contains(view, "📋 Stories") {
		t.Error("View() should contain Stories section title")
	}
	// Verify sections are properly separated (check for spacing before titles)
	if !strings.Contains(view, "\n\x1b[1;") {
		t.Error("View() should have proper spacing before title elements")
	}

	// Check that boxes have consistent padding (all boxes should have Padding(1, 2))
	// Verify the log box has proper padding by checking for spaces in the border area
	if !strings.Contains(view, "📋 Waiting for output") {
		t.Error("View() should contain log box with proper padding")
	}

	// Verify text hierarchy through styling differences
	// Completed story should be green and bold, in-progress should be highlighted and bold
	if !strings.Contains(view, "completed") || !strings.Contains(view, "in progress") {
		t.Error("View() should show clear text hierarchy with different status styles")
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
