package workflow

import "encoding/json"

// PRDReviewVerdict is the agent's self-review verdict on the generated PRD.
type PRDReviewVerdict struct {
	Approved bool   `json:"approved"`
	Summary  string `json:"summary"`
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
