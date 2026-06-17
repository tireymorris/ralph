package runs

import (
	"bufio"
	"encoding/json"
	"os"

	"ralph/internal/shared/runpaths"
	"ralph/internal/shared/runstate"
)

const eventTypeClarifyingQuestions = "EventClarifyingQuestions"

// NeedsSessionReattach reports whether a non-terminal run needs an in-memory workflow session.
func NeedsSessionReattach(status string) bool {
	switch status {
	case runstate.StatusWaitingClarify, runstate.StatusRunning, runstate.StatusImplementing:
		return true
	default:
		return false
	}
}

// LastClarifyingQuestions returns questions from the most recent clarifying-questions event.
func LastClarifyingQuestions(workDir, runID string) ([]string, error) {
	path := runpaths.EventsPath(workDir, runID)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var last []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		questions, ok, err := clarifyingQuestionsFromLine(sc.Bytes())
		if err != nil {
			return nil, err
		}
		if ok {
			last = questions
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return last, nil
}

func clarifyingQuestionsFromLine(line []byte) ([]string, bool, error) {
	var env struct {
		Type    string `json:"type"`
		Payload struct {
			Questions []string `json:"Questions"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(line, &env); err != nil {
		return nil, false, nil
	}
	if env.Type != eventTypeClarifyingQuestions {
		return nil, false, nil
	}
	return env.Payload.Questions, true, nil
}
