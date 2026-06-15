package prompt

type ClarifyData struct {
	CodebaseNote  string
	UserPrompt    string
	QuestionsFile string
}

type ClarificationsData struct {
	Clarifications []QuestionAnswer
}

type PRDGenerateData struct {
	UserPrompt       string
	PRDFile          string
	BranchPrefix     string
	ContextGuidance  string
	Clarifications   []QuestionAnswer
}

type PRDCritiqueRevisionData struct {
	UserPrompt string
	PRDFile    string
	Critique   string
}

type PRDClarificationRevisionData struct {
	UserPrompt     string
	PRDFile        string
	Clarifications []QuestionAnswer
}

type PRDSelfReviewData struct {
	UserPrompt string
	PRDFile    string
	Round      int
	MaxRounds  int
	VerdictFile string
}

type StoryImplementData struct {
	StoryID            string
	Title              string
	Description        string
	AcceptanceCriteria string
	FeatureTestSpec    string
	Context            string
	PRDFile            string
	Completed          int
	Total              int
	DependsOn          []string
}

type DiffReviewData struct {
	Context      string
	PRDFile      string
	ChangedFiles []string
}

type RecoveryData struct {
	AgentMarker  string
	Context      string
	ErrorMessage string
	FindingsJSON string
	Escalate     bool
	Reason       RecoveryReason
	Attempt      int
	MaxAttempts  int
	PRDFile      string
}

type CleanupData struct {
	Context      string
	PRDFile      string
	ChangedFiles []string
}

type FollowUpData struct {
	UserMessage    string
	PRDFile        string
	RunTranscript  string
}

type CodebaseContextData struct {
	Context string
}

type ChangedFilesData struct {
	ChangedFiles []string
}
