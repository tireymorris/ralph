package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
)

func prepMainView(m *Model) {
	if m.width <= 0 {
		m.width = 80
	}
	if m.height <= 0 {
		m.height = 24
	}
	m.applyLayout(m.width, m.height)
	m.rebuildMainScrollContent()
}

func TestViewQuitting(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.quitting = true

	view := m.View()
	if !strings.Contains(view, "Goodbye") {
		t.Errorf("View() when quitting should say goodbye, got %q", view)
	}
}

func TestViewPhaseAwaitingPrompt(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "", false, false, false)
	m.phase = PhasePRDGeneration
	m.prompt = "cached prompt"
	prepMainView(m)
	m.phase = PhaseAwaitingPrompt
	m.prompt = ""
	m.promptInput.Blur()
	m.promptInput.SetValue("")

	view := m.View()
	if !strings.Contains(view, "RALPH") {
		t.Error("View() during PhaseAwaitingPrompt should contain RALPH header")
	}
	if !strings.Contains(view, "Awaiting Prompt") {
		t.Error("View() during PhaseAwaitingPrompt should contain phase label")
	}
	if m.promptInput.Placeholder != "Describe what you want to build" {
		t.Fatalf("promptInput.Placeholder = %q, want %q", m.promptInput.Placeholder, "Describe what you want to build")
	}
	if !strings.Contains(view, "enter") || !strings.Contains(view, "ctrl+c") || !strings.Contains(view, "q/") {
		t.Error("View() during PhaseAwaitingPrompt should show enter and quit help")
	}
	if strings.Contains(view, "Generating") {
		t.Error("View() during PhaseAwaitingPrompt should not show PRD generation layout")
	}
}

func TestViewPhaseInit(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test prompt", false, false, false)
	m.phase = PhaseInit
	m.width = 80
	m.height = 24
	prepMainView(m)

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
	prepMainView(m)

	view := m.View()
	if !strings.Contains(view, "test prompt") || !strings.Contains(view, "Generating") {
		t.Error("View() during PRD generation should show prompt and generating message")
	}
}

func TestViewPhasePRDGenerationLongPrompt(t *testing.T) {
	cfg := config.DefaultConfig()
	prompt := "START_" + strings.Repeat("x", 88) + "_END__"
	if len(prompt) != 100 {
		t.Fatalf("prompt length = %d, want 100", len(prompt))
	}
	m := NewModel(cfg, prompt, false, false, false)
	m.phase = PhasePRDGeneration
	m.width = 80
	m.height = 24
	prepMainView(m)

	view := m.View()
	if !strings.Contains(view, prompt[:20]) {
		t.Errorf("View() should contain first 20 chars of prompt, got %q", view)
	}
	if !strings.Contains(view, prompt[len(prompt)-20:]) {
		t.Errorf("View() should contain last 20 chars of prompt, got %q", view)
	}

	wrapWidth := max(20, m.mainPane.Width-10)
	wrapped := wrapText(prompt, wrapWidth)
	segments := strings.Split(wrapped, "\n")
	if len(segments) < 2 {
		t.Fatalf("sanity: prompt should wrap to at least 2 lines at width %d", wrapWidth)
	}
	lineIndex := func(substr string) int {
		for i, line := range strings.Split(view, "\n") {
			if strings.Contains(line, substr) {
				return i
			}
		}
		return -1
	}
	firstLine := lineIndex(segments[0])
	secondLine := lineIndex(segments[1])
	if firstLine < 0 || secondLine < 0 {
		t.Fatalf("View() missing wrapped prompt segments (first=%d second=%d)", firstLine, secondLine)
	}
	if firstLine == secondLine {
		t.Errorf("View() should wrap prompt across lines, both segments on line %d", firstLine)
	}
}

func TestViewPhaseFailed(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhaseFailed
	m.err = errors.New("AI completed but did not generate prd.json")
	m.width = 80
	m.height = 24
	prepMainView(m)

	view := m.View()
	if !strings.Contains(view, "prd.json") || !strings.Contains(view, "Failed") {
		t.Errorf("View() should show failure phase and error, got %q", view)
	}
	if !strings.Contains(view, "r retry") {
		t.Errorf("View() should mention r retry, got %q", view)
	}
}

func TestViewPhaseFailedLongError(t *testing.T) {
	cfg := config.DefaultConfig()
	errMsg := "START_" + strings.Repeat("x", 78) + "_END__"
	if len(errMsg) != 90 {
		t.Fatalf("error length = %d, want 90", len(errMsg))
	}
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhaseFailed
	m.err = errors.New(errMsg)
	m.width = 60
	m.height = 24
	prepMainView(m)

	view := m.View()
	if !strings.Contains(view, errMsg[:20]) {
		t.Errorf("View() should contain first 20 chars of error, got %q", view)
	}
	if !strings.Contains(view, errMsg[len(errMsg)-20:]) {
		t.Errorf("View() should contain last 20 chars of error, got %q", view)
	}
	if strings.Contains(view, "...") {
		t.Errorf("View() should not truncate error with ellipsis, got %q", view)
	}

	wrapWidth := max(20, m.mainPane.Width-10)
	wrapped := wrapText(errMsg, wrapWidth)
	segments := strings.Split(wrapped, "\n")
	if len(segments) < 2 {
		t.Fatalf("sanity: error should wrap to at least 2 lines at width %d", wrapWidth)
	}
	lineIndex := func(substr string) int {
		for i, line := range strings.Split(view, "\n") {
			if strings.Contains(line, substr) {
				return i
			}
		}
		return -1
	}
	firstLine := lineIndex(segments[0])
	secondLine := lineIndex(segments[1])
	if firstLine < 0 || secondLine < 0 {
		t.Fatalf("View() missing wrapped error segments (first=%d second=%d)", firstLine, secondLine)
	}
	if firstLine == secondLine {
		t.Errorf("View() should wrap error across lines, both segments on line %d", firstLine)
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
	m.height = 45
	prepMainView(m)

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
	prepMainView(m)

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
	prepMainView(m)

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
	m.width = 80
	m.height = 24
	prepMainView(m)

	view := m.View()
	if !strings.Contains(view, "Done Project") {
		t.Error("View() should show project name")
	}
	if !strings.Contains(view, "completed") {
		t.Error("View() should mention completed")
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

	lipgloss.SetColorProfile(termenv.TrueColor)

	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)

	header := m.renderHeader()

	expectedEscape := "\x1b[48;2;168;85;247m"
	if !strings.Contains(header, expectedEscape) {
		t.Errorf("renderHeader() should contain ANSI escape for primary color #A855F7, got: %q", header)
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

func TestRenderLogsStyling(t *testing.T) {

	lipgloss.SetColorProfile(termenv.TrueColor)

	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.width = 80

	logs := m.renderLogs()

	if !strings.Contains(logs, "╭") {
		t.Error("renderLogs() output should contain rounded border characters")
	}

	if !strings.Contains(logs, "\x1b[48;2;17;24;39m") {
		t.Error("renderLogs() output should contain surface color background ANSI sequence")
	}
}

func TestViewLogsPaneShowsOutputLogs(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhaseImplementation
	m.prd = &prd.PRD{
		ProjectName: "Test",
		Stories:     []*prd.Story{{ID: "1", Title: "Story", Passes: false}},
	}
	m.currentStory = m.prd.Stories[0]
	m.scrollPane = focusLogs
	m.width = 80
	m.height = 45
	prepMainView(m)

	view := m.View()
	if !strings.Contains(view, "Output Logs") {
		t.Error("View() should contain 'Output Logs' when logs pane is active")
	}
}

func TestViewMainPaneHidesOutputLogs(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhaseImplementation
	m.prd = &prd.PRD{
		ProjectName: "Test",
		Stories:     []*prd.Story{{ID: "1", Title: "Story", Passes: false}},
	}
	m.currentStory = m.prd.Stories[0]
	m.scrollPane = focusMain
	m.width = 80
	m.height = 45
	prepMainView(m)

	view := m.View()
	if strings.Contains(view, "Output Logs") {
		t.Error("View() should not contain 'Output Logs' when main pane is active")
	}
}

func TestViewPRDReviewShowsCritiqueShortcut(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhasePRDReview
	m.prd = &prd.PRD{
		ProjectName: "Test",
		Stories:     []*prd.Story{{ID: "1", Title: "Story", Passes: false}},
	}
	m.width = 80
	m.height = 40
	prepMainView(m)

	view := m.View()
	if !strings.Contains(view, "Press c to add critique or Enter to continue to implementation") {
		t.Errorf("View() should include critique shortcut in PRD review help, got %q", view)
	}
	if !strings.Contains(view, "c critique") {
		t.Errorf("View() footer help should include critique shortcut, got %q", view)
	}
}

func TestViewPhaseClarifyingLongQuestion(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	question := "START_" + strings.Repeat("x", 138) + "_END__"
	if len(question) != 150 {
		t.Fatalf("question length = %d, want 150", len(question))
	}
	m.phase = PhaseClarifying
	m.clarifyQuestions = []string{question}
	m.width = 80
	m.height = 40
	prepMainView(m)

	view := m.View()
	if !strings.Contains(view, question[:20]) {
		t.Errorf("View() should contain first 20 chars of question, got %q", view)
	}
	if !strings.Contains(view, question[len(question)-20:]) {
		t.Errorf("View() should contain last 20 chars of question, got %q", view)
	}

	contentWidth := max(20, m.width-4)
	wrapped := wrapText(question, contentWidth)
	segments := strings.Split(wrapped, "\n")
	if len(segments) < 2 {
		t.Fatalf("sanity: question should wrap to at least 2 lines at width %d", contentWidth)
	}
	lineIndex := func(substr string) int {
		for i, line := range strings.Split(view, "\n") {
			if strings.Contains(line, substr) {
				return i
			}
		}
		return -1
	}
	firstLine := lineIndex(segments[0])
	secondLine := lineIndex(segments[1])
	if firstLine < 0 || secondLine < 0 {
		t.Fatalf("View() missing wrapped question segments (first=%d second=%d)", firstLine, secondLine)
	}
	if firstLine == secondLine {
		t.Errorf("View() should wrap question across lines, both segments on line %d", firstLine)
	}
}

func TestViewPhaseClarifyingInstructionWrap(t *testing.T) {
	const instruction = "Please answer the following questions before we generate your PRD."

	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhaseClarifying
	m.clarifyQuestions = []string{"Short question?"}
	m.width = 50
	m.height = 40
	prepMainView(m)

	contentWidth := max(20, m.width-4)
	wrapped := wrapText(instruction, contentWidth)
	segments := strings.Split(wrapped, "\n")
	if len(segments) < 2 {
		t.Fatalf("sanity: instruction should wrap to at least 2 lines at width %d", contentWidth)
	}

	view := m.View()
	lineIndex := func(substr string) int {
		for i, line := range strings.Split(view, "\n") {
			if strings.Contains(line, substr) {
				return i
			}
		}
		return -1
	}
	firstLine := lineIndex(segments[0])
	secondLine := lineIndex(segments[1])
	if firstLine < 0 || secondLine < 0 {
		t.Fatalf("View() missing wrapped instruction segments (first=%d second=%d)", firstLine, secondLine)
	}
	if firstLine == secondLine {
		t.Errorf("View() should wrap instruction across lines, both segments on line %d", firstLine)
	}

	const navHint = "  Tab/↑/↓ to navigate  •  Enter to confirm  •  Esc to skip all questions"
	wrappedNav := wrapText(navHint, contentWidth)
	navSegments := strings.Split(wrappedNav, "\n")
	if len(navSegments) < 2 {
		t.Fatalf("sanity: navigation hint should wrap to at least 2 lines at width %d", contentWidth)
	}
	navFirst := lineIndex(navSegments[0])
	navSecond := lineIndex(navSegments[1])
	if navFirst < 0 || navSecond < 0 {
		t.Fatalf("View() missing wrapped navigation segments (first=%d second=%d)", navFirst, navSecond)
	}
	if navFirst == navSecond {
		t.Errorf("View() should wrap navigation hint across lines, both segments on line %d", navFirst)
	}
}

func TestViewPhasePRDReviewLongAcceptanceCriteria(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	criterion := "START_" + strings.Repeat("x", 108) + "_END__"
	if len(criterion) != 120 {
		t.Fatalf("criterion length = %d, want 120", len(criterion))
	}
	m.phase = PhasePRDReview
	m.prd = &prd.PRD{
		ProjectName: "Test",
		Stories: []*prd.Story{{
			ID:                 "1",
			Title:              "Story",
			Passes:             false,
			AcceptanceCriteria: []string{criterion},
		}},
	}
	m.width = 80
	m.height = 40
	prepMainView(m)

	view := m.View()
	lineIndex := func(substr string) int {
		for i, line := range strings.Split(view, "\n") {
			if strings.Contains(line, substr) {
				return i
			}
		}
		return -1
	}
	firstLine := lineIndex(criterion[:20])
	secondLine := lineIndex(criterion[len(criterion)-20:])
	if firstLine < 0 || secondLine < 0 {
		t.Fatalf("View() missing criterion text (first=%d second=%d)", firstLine, secondLine)
	}
	if firstLine == secondLine {
		t.Errorf("View() should wrap acceptance criterion across lines, both segments on line %d", firstLine)
	}
	if !strings.Contains(strings.Split(view, "\n")[firstLine], "      - ") {
		t.Errorf("first criterion line should start with indent prefix, got %q", strings.Split(view, "\n")[firstLine])
	}
}

func TestViewPRDReviewShowsCritiqueInputWhenActive(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false, false)
	m.phase = PhasePRDReview
	m.prd = &prd.PRD{
		ProjectName: "Test",
		Stories:     []*prd.Story{{ID: "1", Title: "Story", Passes: false}},
	}
	m.critiqueActive = true
	m.critiqueInput.SetValue("Needs better tests")
	m.width = 80
	m.height = 40
	prepMainView(m)

	view := m.View()
	if !strings.Contains(view, "Critique (Enter submit • Esc cancel)") {
		t.Errorf("View() should show critique input help when critique mode is active, got %q", view)
	}
	if !strings.Contains(view, "Needs better tests") {
		t.Errorf("View() should include critique input value when active, got %q", view)
	}
}
