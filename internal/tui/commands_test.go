package tui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ralph/internal/config"
	"ralph/internal/prd"
	"ralph/internal/runner"
)

type mockGenerator struct {
	prd *prd.PRD
	err error
}

func (m *mockGenerator) Generate(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) (*prd.PRD, error) {
	return m.prd, m.err
}

type mockImplementer struct {
	success bool
	err     error
}

func (m *mockImplementer) Implement(ctx context.Context, story *prd.Story, iteration int, p *prd.PRD, outputCh chan<- runner.OutputLine) (bool, error) {
	return m.success, m.err
}

func TestLoadAndResumeError(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.PRDFile = "/nonexistent/path/prd.json"

	m := NewModel(cfg, "", false, true)
	msg := m.loadAndResume()

	if _, ok := msg.(prdErrorMsg); !ok {
		t.Errorf("loadAndResume() should return prdErrorMsg, got %T", msg)
	}
}

func TestLoadAndResumeSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	prdFile := filepath.Join(tmpDir, "test.json")

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories:     []*prd.Story{{ID: "1", Title: "T", Description: "D", AcceptanceCriteria: []string{"a"}, Priority: 1}},
	}

	cfg := &config.Config{PRDFile: prdFile}
	prd.Save(cfg, testPRD)

	m := NewModel(cfg, "", false, true)
	msg := m.loadAndResume()

	if genMsg, ok := msg.(prdGeneratedMsg); !ok {
		t.Errorf("loadAndResume() should return prdGeneratedMsg, got %T", msg)
	} else if genMsg.prd.ProjectName != "Test" {
		t.Errorf("loadAndResume() PRD name = %q, want %q", genMsg.prd.ProjectName, "Test")
	}
}

func TestStartNextStoryAllCompleted(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false)
	m.prd = &prd.PRD{
		Stories: []*prd.Story{
			{ID: "1", Passes: true},
		},
	}

	msg := m.startNextStory()

	if phase, ok := msg.(phaseChangeMsg); !ok {
		t.Errorf("startNextStory() should return phaseChangeMsg, got %T", msg)
	} else if Phase(phase) != PhaseCompleted {
		t.Errorf("startNextStory() phase = %v, want PhaseCompleted", phase)
	}
}

func TestStartNextStoryAllFailed(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RetryAttempts = 3
	m := NewModel(cfg, "test", false, false)
	m.prd = &prd.PRD{
		Stories: []*prd.Story{
			{ID: "1", Passes: false, RetryCount: 5},
		},
	}

	msg := m.startNextStory()

	if phase, ok := msg.(phaseChangeMsg); !ok {
		t.Errorf("startNextStory() should return phaseChangeMsg, got %T", msg)
	} else if Phase(phase) != PhaseFailed {
		t.Errorf("startNextStory() phase = %v, want PhaseFailed", phase)
	}
}

func TestListenForOutputContextDone(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false)

	m.cancelFunc()

	cmd := m.listenForOutput()
	msg := cmd()

	if msg != nil {
		t.Errorf("listenForOutput() with cancelled context should return nil, got %T", msg)
	}
}

func TestListenForOutputChannelClosed(t *testing.T) {
	cfg := config.DefaultConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := &Model{
		cfg:      cfg,
		ctx:      ctx,
		outputCh: make(chan runner.OutputLine, 10),
	}
	close(m.outputCh)

	cmd := m.listenForOutput()
	msg := cmd()

	if msg != nil {
		t.Errorf("listenForOutput() with closed channel should return nil, got %T", msg)
	}
}

func TestListenForOutputNormalMessage(t *testing.T) {
	cfg := config.DefaultConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := &Model{
		cfg:      cfg,
		ctx:      ctx,
		outputCh: make(chan runner.OutputLine, 10),
	}

	m.outputCh <- runner.OutputLine{Text: "test output", IsErr: false}

	cmd := m.listenForOutput()
	msg := cmd()

	if output, ok := msg.(outputMsg); !ok {
		t.Errorf("listenForOutput() should return outputMsg, got %T", msg)
	} else if output.Text != "test output" {
		t.Errorf("output.Text = %q, want %q", output.Text, "test output")
	}
}

func TestListenForOutputStoryCompleteSuccess(t *testing.T) {
	cfg := config.DefaultConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := &Model{
		cfg:      cfg,
		ctx:      ctx,
		outputCh: make(chan runner.OutputLine, 10),
	}

	m.outputCh <- runner.OutputLine{Text: "STORY_COMPLETE:success"}

	cmd := m.listenForOutput()
	msg := cmd()

	if complete, ok := msg.(storyCompleteMsg); !ok {
		t.Errorf("listenForOutput() should return storyCompleteMsg, got %T", msg)
	} else if !complete.success {
		t.Error("storyCompleteMsg.success should be true")
	}
}

func TestListenForOutputStoryCompleteFailure(t *testing.T) {
	cfg := config.DefaultConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := &Model{
		cfg:      cfg,
		ctx:      ctx,
		outputCh: make(chan runner.OutputLine, 10),
	}

	m.outputCh <- runner.OutputLine{Text: "STORY_COMPLETE:failure"}

	cmd := m.listenForOutput()
	msg := cmd()

	if complete, ok := msg.(storyCompleteMsg); !ok {
		t.Errorf("listenForOutput() should return storyCompleteMsg, got %T", msg)
	} else if complete.success {
		t.Error("storyCompleteMsg.success should be false")
	}
}

func TestContinueImplementationAllCompleted(t *testing.T) {
	tmpDir := t.TempDir()
	prdFile := filepath.Join(tmpDir, "prd.json")

	cfg := config.DefaultConfig()
	cfg.PRDFile = prdFile

	testPRD := &prd.PRD{
		Stories: []*prd.Story{{ID: "1", Passes: true}},
	}
	prd.Save(cfg, testPRD)

	m := NewModel(cfg, "test", false, false)
	m.prd = testPRD

	cmd := m.continueImplementation()
	msg := cmd()

	if phase, ok := msg.(phaseChangeMsg); !ok {
		t.Errorf("continueImplementation() should return phaseChangeMsg, got %T", msg)
	} else if Phase(phase) != PhaseCompleted {
		t.Errorf("phase = %v, want PhaseCompleted", phase)
	}
}

func TestContinueImplementationNoNextStory(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RetryAttempts = 3

	m := NewModel(cfg, "test", false, false)
	m.prd = &prd.PRD{
		Stories: []*prd.Story{{ID: "1", Passes: false, RetryCount: 5}},
	}

	cmd := m.continueImplementation()
	msg := cmd()

	if phase, ok := msg.(phaseChangeMsg); !ok {
		t.Errorf("continueImplementation() should return phaseChangeMsg, got %T", msg)
	} else if Phase(phase) != PhaseFailed {
		t.Errorf("phase = %v, want PhaseFailed", phase)
	}
}

func TestContinueImplementationMaxIterations(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.MaxIterations = 5

	m := NewModel(cfg, "test", false, false)
	m.prd = &prd.PRD{
		Stories: []*prd.Story{{ID: "1", Passes: false}},
	}
	m.iteration = 10

	cmd := m.continueImplementation()
	msg := cmd()

	if phase, ok := msg.(phaseChangeMsg); !ok {
		t.Errorf("continueImplementation() should return phaseChangeMsg, got %T", msg)
	} else if Phase(phase) != PhaseFailed {
		t.Errorf("phase = %v, want PhaseFailed", phase)
	}
}

func TestSetupBranchAndStartNoBranch(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RetryAttempts = 3

	m := NewModel(cfg, "test", false, false)
	m.prd = &prd.PRD{
		BranchName: "",
		Stories:    []*prd.Story{{ID: "1", Passes: true}},
	}

	cmd := m.setupBranchAndStart()

	done := make(chan struct{})
	go func() {
		cmd()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("setupBranchAndStart() timed out")
	}
}

func TestStartOperationResume(t *testing.T) {
	tmpDir := t.TempDir()
	prdFile := filepath.Join(tmpDir, "prd.json")

	cfg := config.DefaultConfig()
	cfg.PRDFile = prdFile

	testPRD := &prd.PRD{ProjectName: "Resume Test", Stories: []*prd.Story{{ID: "1", Title: "T", Description: "D", AcceptanceCriteria: []string{"a"}, Priority: 1}}}
	prd.Save(cfg, testPRD)

	m := NewModel(cfg, "", false, true)

	// startOperation now returns phase change first
	cmd := m.startOperation()
	msg := cmd()

	if phaseMsg, ok := msg.(phaseChangeMsg); !ok {
		t.Errorf("startOperation() should return phaseChangeMsg, got %T", msg)
	} else if Phase(phaseMsg) != PhasePRDGeneration {
		t.Errorf("phase = %v, want PhasePRDGeneration", phaseMsg)
	}

	// Then runPRDOperation does the actual load
	cmd = m.runPRDOperation()
	msg = cmd()

	if genMsg, ok := msg.(prdGeneratedMsg); !ok {
		t.Errorf("runPRDOperation() with resume should return prdGeneratedMsg, got %T", msg)
	} else if genMsg.prd.ProjectName != "Resume Test" {
		t.Errorf("PRD name = %q, want %q", genMsg.prd.ProjectName, "Resume Test")
	}
}

func TestStartOperationResumeError(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.PRDFile = "/nonexistent/prd.json"

	m := NewModel(cfg, "", false, true)

	// startOperation returns phase change
	cmd := m.startOperation()
	msg := cmd()

	if _, ok := msg.(phaseChangeMsg); !ok {
		t.Errorf("startOperation() should return phaseChangeMsg, got %T", msg)
	}

	// runPRDOperation returns the error
	cmd = m.runPRDOperation()
	msg = cmd()

	if _, ok := msg.(prdErrorMsg); !ok {
		t.Errorf("runPRDOperation() with missing file should return prdErrorMsg, got %T", msg)
	}
}

func TestMessageTypes(t *testing.T) {
	_ = outputMsg{}
	_ = prdGeneratedMsg{prd: nil}
	_ = prdErrorMsg{err: nil}
	_ = storyStartMsg{story: nil}
	_ = storyCompleteMsg{success: true}
	_ = storyErrorMsg{err: nil}
	_ = phaseChangeMsg(PhaseInit)
	_ = tickMsg{}
}

func TestGeneratePRDSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.PRDFile = filepath.Join(tmpDir, "prd.json")

	m := NewModel(cfg, "test prompt", false, false)
	testPRD := &prd.PRD{ProjectName: "Test", Stories: []*prd.Story{{ID: "1", Title: "T", Description: "D", AcceptanceCriteria: []string{"a"}, Priority: 1}}}
	m.SetGenerator(&mockGenerator{prd: testPRD})

	msg := m.generatePRD()

	if genMsg, ok := msg.(prdGeneratedMsg); !ok {
		t.Errorf("generatePRD() should return prdGeneratedMsg, got %T", msg)
	} else if genMsg.prd.ProjectName != "Test" {
		t.Errorf("ProjectName = %q, want %q", genMsg.prd.ProjectName, "Test")
	}
}

func TestGeneratePRDGeneratorError(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false)
	m.SetGenerator(&mockGenerator{err: errors.New("gen error")})

	msg := m.generatePRD()

	if _, ok := msg.(prdErrorMsg); !ok {
		t.Errorf("generatePRD() should return prdErrorMsg on error, got %T", msg)
	}
}

func TestGeneratePRDSaveError(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.PRDFile = "/nonexistent/dir/prd.json"

	m := NewModel(cfg, "test", false, false)
	testPRD := &prd.PRD{ProjectName: "Test", Stories: []*prd.Story{{ID: "1"}}}
	m.SetGenerator(&mockGenerator{prd: testPRD})

	msg := m.generatePRD()

	if _, ok := msg.(prdErrorMsg); !ok {
		t.Errorf("generatePRD() should return prdErrorMsg on save error, got %T", msg)
	}
}

func TestStartNextStoryWithImplementer(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RetryAttempts = 3

	m := NewModel(cfg, "test", false, false)
	m.prd = &prd.PRD{
		Stories: []*prd.Story{{ID: "1", Passes: false}},
	}
	m.SetImplementer(&mockImplementer{success: true})

	msg := m.startNextStory()

	if startMsg, ok := msg.(storyStartMsg); !ok {
		t.Errorf("startNextStory() should return storyStartMsg, got %T", msg)
	} else if startMsg.story.ID != "1" {
		t.Errorf("story.ID = %q, want %q", startMsg.story.ID, "1")
	}

	time.Sleep(50 * time.Millisecond)
}

func TestStartNextStoryWithImplementerError(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RetryAttempts = 3

	m := NewModel(cfg, "test", false, false)
	m.prd = &prd.PRD{
		Stories: []*prd.Story{{ID: "1", Passes: false}},
	}
	m.SetImplementer(&mockImplementer{err: errors.New("impl error")})

	msg := m.startNextStory()

	if _, ok := msg.(storyStartMsg); !ok {
		t.Errorf("startNextStory() should return storyStartMsg, got %T", msg)
	}

	time.Sleep(50 * time.Millisecond)
}

func TestStartNextStoryWithImplementerFailure(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RetryAttempts = 3

	m := NewModel(cfg, "test", false, false)
	m.prd = &prd.PRD{
		Stories: []*prd.Story{{ID: "1", Passes: false}},
	}
	m.SetImplementer(&mockImplementer{success: false})

	msg := m.startNextStory()

	if _, ok := msg.(storyStartMsg); !ok {
		t.Errorf("startNextStory() should return storyStartMsg, got %T", msg)
	}

	time.Sleep(50 * time.Millisecond)
}

func TestStartOperationGenerate(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.PRDFile = filepath.Join(tmpDir, "prd.json")

	m := NewModel(cfg, "test prompt", false, false)
	testPRD := &prd.PRD{ProjectName: "Test", Stories: []*prd.Story{{ID: "1"}}}
	m.SetGenerator(&mockGenerator{prd: testPRD})

	// startOperation returns phase change first
	cmd := m.startOperation()
	msg := cmd()

	if phaseMsg, ok := msg.(phaseChangeMsg); !ok {
		t.Errorf("startOperation() should return phaseChangeMsg, got %T", msg)
	} else if Phase(phaseMsg) != PhasePRDGeneration {
		t.Errorf("phase = %v, want PhasePRDGeneration", phaseMsg)
	}

	// runPRDOperation does the actual generation
	cmd = m.runPRDOperation()
	msg = cmd()

	if _, ok := msg.(prdGeneratedMsg); !ok {
		t.Errorf("runPRDOperation() generate path should return prdGeneratedMsg, got %T", msg)
	}
}

func TestSetGenerator(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false)
	g := &mockGenerator{}

	m.SetGenerator(g)

	if m.generator != g {
		t.Error("SetGenerator() did not set generator")
	}
}

func TestSetImplementer(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg, "test", false, false)
	i := &mockImplementer{}

	m.SetImplementer(i)

	if m.implementer != i {
		t.Error("SetImplementer() did not set implementer")
	}
}

func TestContinueImplementationContinue(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RetryAttempts = 3
	cfg.MaxIterations = 100

	m := NewModel(cfg, "test", false, false)
	m.prd = &prd.PRD{
		Stories: []*prd.Story{
			{ID: "1", Passes: false},
			{ID: "2", Passes: true},
		},
	}
	m.iteration = 1
	m.SetImplementer(&mockImplementer{success: true})

	cmd := m.continueImplementation()
	msg := cmd()

	if _, ok := msg.(storyStartMsg); !ok {
		t.Errorf("continueImplementation() should return storyStartMsg when continuing, got %T", msg)
	}

	time.Sleep(50 * time.Millisecond)
}

func TestSetupBranchAndStartWithBranch(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	cfg := config.DefaultConfig()
	cfg.RetryAttempts = 3

	m := NewModel(cfg, "test", false, false)
	m.prd = &prd.PRD{
		BranchName: "feature/test",
		Stories:    []*prd.Story{{ID: "1", Passes: true}},
	}
	m.logs = []string{}

	cmd := m.setupBranchAndStart()

	done := make(chan struct{})
	go func() {
		cmd()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("setupBranchAndStart() timed out")
	}
}
