package workflow

import (
	"fmt"
	"os"
	"path/filepath"
)

// ClarifyingQuestionsFile is the temporary JSON file the AI writes questions to.
const ClarifyingQuestionsFile = ".ralph/questions.json"

// ensureStateDir creates .ralph/ before an agent is asked to write files into it.
func ensureStateDir(workDir string) error {
	return os.MkdirAll(filepath.Join(workDir, ".ralph"), 0755)
}

type QuestionsFileReader struct {
	WorkDir string
}

// ReadRemove reads the questions file and deletes it.
func (q QuestionsFileReader) ReadRemove() ([]byte, error) {
	path := filepath.Join(q.WorkDir, ClarifyingQuestionsFile)
	data, err := os.ReadFile(path)
	if removeErr := os.Remove(path); removeErr != nil && err == nil && !os.IsNotExist(removeErr) {
		return data, fmt.Errorf("remove clarifying questions file %q: %w", path, removeErr)
	}
	return data, err
}
