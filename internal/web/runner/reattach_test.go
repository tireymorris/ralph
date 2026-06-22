package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/runner"
	runctrl "ralph/internal/web/runner"
	"ralph/internal/web/runs"
	"ralph/internal/workflow/events"
)

func TestReattachWaitingClarifyAcceptsAnswers(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	runID := "run-reattach"
	run := &runs.Run{
		ID:        runID,
		WorkDir:   workDir,
		Prompt:    "build feature",
		Status:    "waiting_clarify",
		Phase:     "clarify",
		PRDPath:   "prd.json",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := reg.Register(run); err != nil {
		t.Fatal(err)
	}

	line, err := runctrl.MarshalEventEnvelope(events.EventClarifyingQuestions{
		Questions: []string{"What scope?"},
	})
	if err != nil {
		t.Fatal(err)
	}
	eventsDir := filepath.Join(workDir, ".ralph", "runs", runID)
	if err := os.MkdirAll(eventsDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(eventsDir, "events.ndjson"), append(line, '\n'), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	ctrl := runctrl.NewControllerWithRunner(cfg, reg, runID, &reattachRunner{workDir: workDir})
	t.Cleanup(ctrl.Cancel)

	ctrl.Reattach(context.Background())
	if !ctrl.WaitingForClarify() {
		t.Fatal("expected controller to be waiting for clarify after reattach")
	}

	if err := ctrl.SubmitClarify([]prompt.QuestionAnswer{{
		Question: "What scope?",
		Answer:   "API only",
	}}); err != nil {
		t.Fatalf("SubmitClarify: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		got, ok := reg.Get(runID)
		if ok && got.Status == "waiting_review" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	got, _ := reg.Get(runID)
	t.Fatalf("run status = %q phase = %q, want waiting_review", got.Status, got.Phase)
}

type reattachRunner struct {
	workDir string
}

func (r *reattachRunner) Run(ctx context.Context, prompt string, _ chan<- runner.OutputLine) error {
	if strings.Contains(prompt, "implementation agent") {
		<-ctx.Done()
		return ctx.Err()
	}
	prdPath := filepath.Join(r.workDir, "prd.json")
	data := `{"version":1,"project_name":"Test","branch_name":"feature/x","stories":[{"id":"s1","title":"Story","description":"Do it","slices":[{"id":"slice-1","behavior":"AC","red_hint":"add failing test","passes":false}],"priority":1,"passes":false}]}`
	return os.WriteFile(prdPath, []byte(data), 0644)
}

func (r *reattachRunner) RunnerName() string        { return "mock" }
func (r *reattachRunner) CommandName() string       { return "mock" }
func (r *reattachRunner) IsInternalLog(string) bool { return false }
