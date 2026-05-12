package workflow

import (
	"fmt"
	"os"
	"path/filepath"
)

// ClarifyingQuestionsFile is the temporary JSON file the AI writes clarifying questions to.
const ClarifyingQuestionsFile = ".ralph_questions.json"

type QuestionsFileReader struct {
	WorkDir string
}

// ReadRemove reads the clarifying questions file from WorkDir and deletes it.
func (q QuestionsFileReader) ReadRemove() ([]byte, error) {
	path := filepath.Join(q.WorkDir, ClarifyingQuestionsFile)
	data, err := os.ReadFile(path)
	if removeErr := os.Remove(path); removeErr != nil && err == nil && !os.IsNotExist(removeErr) {
		return data, fmt.Errorf("remove clarifying questions file %q: %w", path, removeErr)
	}
	return data, err
}
