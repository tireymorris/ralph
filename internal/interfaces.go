package internal

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"ralph/internal/prd"
	"ralph/internal/runner"
)

// WorkflowRunner defines the interface for executing workflows
type WorkflowRunner interface {
	RunGenerate(ctx context.Context, prompt string) (*prd.PRD, error)
	RunLoad(ctx context.Context) (*prd.PRD, error)
	RunImplementation(ctx context.Context, p *prd.PRD) error
}

// PRDGenerator defines the interface for generating PRDs
type PRDGenerator interface {
	Generate(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) (*prd.PRD, error)
}

// StoryImplementer defines the interface for implementing stories
type StoryImplementer interface {
	Implement(ctx context.Context, story *prd.Story, iteration int, p *prd.PRD, outputCh chan<- runner.OutputLine) (bool, error)
}

// GitManager defines the interface for git operations
type GitManager interface {
	IsRepository() bool
	CurrentBranch() (string, error)
	BranchExists(name string) bool
	CreateBranch(name string) error
	Checkout(name string) error
	HasChanges() bool
	StageAll() error
	Commit(message string) error
	CommitStory(storyID, title, description string) error
}

// TUIModel defines the interface for TUI models
type TUIModel interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
	SetGenerator(g PRDGenerator)
	SetImplementer(i StoryImplementer)
}

// GitCommitter defines the interface for committing git changes for stories
type GitCommitter interface {
	CommitStory(storyID, title, description string) error
}

// PRDStorage defines the interface for PRD persistence
type PRDStorage interface {
	Load() (*prd.PRD, error)
	Save(p *prd.PRD) error
	Delete() error
}
