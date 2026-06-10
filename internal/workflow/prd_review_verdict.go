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

type verdictReadStatus int

const (
	verdictOK verdictReadStatus = iota
	verdictMissing
	verdictMalformed
)

type PRDReviewVerdictReader struct {
	WorkDir string
}

func (r PRDReviewVerdictReader) Path() string {
	return filepath.Join(r.WorkDir, prompt.PRDSelfReviewVerdictFile)
}

// ReadRemove reads the verdict file and deletes it. A missing file is not an
// error and yields a zero-value verdict for callers that treat absence as
// "not approved".
func (r PRDReviewVerdictReader) ReadRemove() (PRDReviewVerdict, error) {
	v, status, err := r.readAndRemove()
	if err != nil {
		return PRDReviewVerdict{}, err
	}
	if status == verdictMissing || status == verdictMalformed {
		return PRDReviewVerdict{}, nil
	}
	return v, nil
}

func (r PRDReviewVerdictReader) readAndRemove() (PRDReviewVerdict, verdictReadStatus, error) {
	path := r.Path()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return PRDReviewVerdict{}, verdictMissing, nil
		}
		return PRDReviewVerdict{}, verdictOK, fmt.Errorf("read prd review verdict file %q: %w", path, err)
	}
	if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
		return PRDReviewVerdict{}, verdictOK, fmt.Errorf("remove prd review verdict file %q: %w", path, removeErr)
	}
	if len(data) == 0 || !json.Valid(data) {
		return PRDReviewVerdict{}, verdictMalformed, nil
	}
	var v PRDReviewVerdict
	if err := json.Unmarshal(data, &v); err != nil {
		return PRDReviewVerdict{}, verdictMalformed, nil
	}
	return v, verdictOK, nil
}

func parsePRDReviewVerdict(data []byte) PRDReviewVerdict {
	if len(data) == 0 || !json.Valid(data) {
		return PRDReviewVerdict{}
	}
	var v PRDReviewVerdict
	if err := json.Unmarshal(data, &v); err != nil {
		return PRDReviewVerdict{}
	}
	return v
}
