package tui

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"ralph/internal/config"
	"ralph/internal/prd"
	"ralph/internal/prompt"
)

// Phase is the Bubble Tea UI phase (distinct from workflow.Executor phases).
type Phase int

const (
	PhaseInit Phase = iota
	PhaseClarifying
	PhasePRDGeneration
	PhasePRDReview
	PhaseImplementation
	PhaseCompleted
)

func (p Phase) String() string {
	switch p {
	case PhaseInit:
		return "Initializing"
	case PhaseClarifying:
		return "Clarifying Questions"
	case PhasePRDGeneration:
		return "Phase 1: PRD Generation"
	case PhasePRDReview:
		return "PRD Review"
	case PhaseImplementation:
		return "Phase 2: Implementation"
	case PhaseCompleted:
		return "Completed"
	default:
		return "Unknown"
	}
}

// Model is the root Bubble Tea model for interactive Ralph.
type Model struct {
	cfg     *config.Config
	prompt  string
	dryRun  bool
	resume  bool
	verbose bool

	phase        Phase
	prd          *prd.PRD
	currentStory *prd.Story
	err          error
	quitting     bool
	width        int
	height       int

	spinner  spinner.Model
	progress progress.Model

	mainPane      viewport.Model
	scrollPane    scrollFocus
	fullscreenPane scrollFocus
	snapMainToTop bool

	logHeightBias int

	layoutSigW, layoutSigH int
	layoutSigLogCount      int
	layoutSigBias          int

	clarifyQuestions []string
	clarifyInputs    []textinput.Model
	clarifyFocused   int
	clarifyAnswersCh chan<- []prompt.QuestionAnswer

	logger           *Logger
	operationManager *OperationManager
}

func NewModel(cfg *config.Config, prompt string, dryRun, resume, verbose bool) *Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(accentColor)

	p := progress.New(
		progress.WithGradient("#A855F7", "#10B981"),
		progress.WithWidth(40),
		progress.WithSolidFill("#4B5563"),
	)

	mv := viewport.New(80, 12)
	mv.Style = lipgloss.NewStyle()
	mv.MouseWheelEnabled = true

	logger := NewLogger(verbose)
	operationManager := NewOperationManager(cfg)

	return &Model{
		cfg:              cfg,
		prompt:           prompt,
		dryRun:           dryRun,
		resume:           resume,
		verbose:          verbose,
		phase:            PhaseInit,
		spinner:          s,
		progress:         p,
		mainPane:         mv,
		scrollPane:       focusMain,
		logger:           logger,
		operationManager: operationManager,
	}
}

func (m *Model) ExitCode() int {
	if m.phase == PhaseCompleted {
		return 0
	}
	return 1
}
