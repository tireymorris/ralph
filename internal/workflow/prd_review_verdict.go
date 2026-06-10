package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"ralph/internal/prompt"
)

// PRDReviewVerdict is the agent's self-review verdict on the generated PRD.
type PRDReviewVerdict struct {
	Approved bool   `json:"approved"`
	Summary  string `json:"summary"`
}

type PRDReviewVerdictReader struct {
	WorkDir string
}

func (r PRDReviewVerdictReader) Path() string {
	return filepath.Join(r.WorkDir, prompt.PRDSelfReviewVerdictFile)
}

// ReadRemove reads the verdict file and deletes it. Unlike QuestionsFileReader,
// a missing file is not an error: it yields a zero-value verdict so a
// non-cooperating runner cannot fail the run.
func (r PRDReviewVerdictReader) ReadRemove() (PRDReviewVerdict, error) {
	path := r.Path()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return PRDReviewVerdict{}, nil
		}
		return PRDReviewVerdict{}, fmt.Errorf("read prd review verdict file %q: %w", path, err)
	}
	if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
		return parsePRDReviewVerdict(data), fmt.Errorf("remove prd review verdict file %q: %w", path, removeErr)
	}
	return parsePRDReviewVerdict(data), nil
}

// parsePRDReviewVerdict treats malformed data as a zero-value verdict:
// a non-cooperating runner must not fail the run.
func parsePRDReviewVerdict(data []byte) PRDReviewVerdict {
	var v PRDReviewVerdict
	if err := json.Unmarshal(data, &v); err != nil {
		return PRDReviewVerdict{}
	}
	return v
}
