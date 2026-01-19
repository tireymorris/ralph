package prd

// Story represents a single user story in the PRD
type Story struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	TestSpec           string   `json:"test_spec,omitempty"`
	Priority           int      `json:"priority"`
	Passes             bool     `json:"passes"`
	RetryCount         int      `json:"retry_count"`
}

// PRD represents the Product Requirements Document
type PRD struct {
	ProjectName string   `json:"project_name"`
	BranchName  string   `json:"branch_name,omitempty"`
	Stories     []*Story `json:"stories"`
}

// StoryStatus represents the current status of a story
type StoryStatus int

const (
	StatusPending StoryStatus = iota
	StatusInProgress
	StatusCompleted
	StatusFailed
)

func (s *Story) Status() StoryStatus {
	if s.Passes {
		return StatusCompleted
	}
	return StatusPending
}
