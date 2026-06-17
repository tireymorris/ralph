package runs_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ralph/internal/shared/runpaths"
	"ralph/internal/shared/runstate"
	"ralph/internal/web/runner"
	"ralph/internal/web/runs"
	"ralph/internal/workflow/events"
)

func TestNeedsSessionReattach(t *testing.T) {
	if !runs.NeedsSessionReattach(runstate.StatusWaitingClarify) {
		t.Fatal("waiting_clarify should need reattach")
	}
	if !runs.NeedsSessionReattach(runstate.StatusRunning) {
		t.Fatal("running should need reattach")
	}
	if runs.NeedsSessionReattach(runstate.StatusWaitingReview) {
		t.Fatal("waiting_review should not need reattach")
	}
}

func TestLastClarifyingQuestions(t *testing.T) {
	workDir := t.TempDir()
	runID := "run-clarify"
	dir := filepath.Join(workDir, ".ralph", "runs", runID)
	if err := os.MkdirAll(dir, 0750); err != nil {
		t.Fatal(err)
	}

	first, err := runner.MarshalEventEnvelope(events.EventOutput{Output: events.Output{Text: "hello"}})
	if err != nil {
		t.Fatal(err)
	}
	second, err := runner.MarshalEventEnvelope(events.EventClarifyingQuestions{
		Questions: []string{"First?"},
	})
	if err != nil {
		t.Fatal(err)
	}
	third, err := runner.MarshalEventEnvelope(events.EventClarifyingQuestions{
		Questions: []string{"Latest?", "Second latest?"},
	})
	if err != nil {
		t.Fatal(err)
	}

	path := runpaths.EventsPath(workDir, runID)
	if err := os.WriteFile(path, []byte(string(first)+"\n"+string(second)+"\n"+string(third)+"\n"), 0600); err != nil {
		t.Fatal(err)
	}

	got, err := runs.LastClarifyingQuestions(workDir, runID)
	if err != nil {
		t.Fatalf("LastClarifyingQuestions: %v", err)
	}
	if len(got) != 2 || got[0] != "Latest?" || got[1] != "Second latest?" {
		t.Fatalf("questions = %#v, want latest pair", got)
	}
}

func TestLastClarifyingQuestionsMissingFile(t *testing.T) {
	got, err := runs.LastClarifyingQuestions(t.TempDir(), "missing")
	if err != nil {
		t.Fatalf("LastClarifyingQuestions: %v", err)
	}
	if got != nil {
		t.Fatalf("questions = %#v, want nil", got)
	}
}

func TestLastClarifyingQuestionsIgnoresInvalidLines(t *testing.T) {
	workDir := t.TempDir()
	runID := "run-parse"
	dir := filepath.Join(workDir, ".ralph", "runs", runID)
	if err := os.MkdirAll(dir, 0750); err != nil {
		t.Fatal(err)
	}

	payload, err := json.Marshal(map[string]any{
		"type": "EventClarifyingQuestions",
		"payload": map[string]any{
			"Questions": []string{"What scope?"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	path := runpaths.EventsPath(workDir, runID)
	if err := os.WriteFile(path, []byte("not json\n"+string(payload)+"\n"), 0600); err != nil {
		t.Fatal(err)
	}

	got, err := runs.LastClarifyingQuestions(workDir, runID)
	if err != nil {
		t.Fatalf("LastClarifyingQuestions: %v", err)
	}
	if len(got) != 1 || got[0] != "What scope?" {
		t.Fatalf("questions = %#v", got)
	}
}
