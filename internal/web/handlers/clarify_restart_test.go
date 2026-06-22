package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/runner"
	"ralph/internal/shared/testgit"
	"ralph/internal/web/handlers"
	webrunner "ralph/internal/web/runner"
	"ralph/internal/web/runs"
	"ralph/internal/workflow/events"
)

func TestClarifySubmitRecreatesControllerAfterRestart(t *testing.T) {
	workDir := t.TempDir()
	testgit.InitRepo(t, workDir)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	reg := runs.NewRegistry()

	runID := "run-clarify-restart"
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

	eventLine, err := webrunner.MarshalEventEnvelope(events.EventClarifyingQuestions{
		Questions: []string{"What scope?"},
	})
	if err != nil {
		t.Fatal(err)
	}
	eventsDir := filepath.Join(workDir, ".ralph", "runs", runID)
	if err := os.MkdirAll(eventsDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(eventsDir, "events.ndjson"), append(eventLine, '\n'), 0600); err != nil {
		t.Fatal(err)
	}

	api := handlers.NewAPI(cfg, reg)
	api.SetRunnerFactory(func(*config.Config) (runner.RunnerInterface, error) {
		return &clarifyRestartRunner{workDir: workDir}, nil
	})
	t.Cleanup(api.ReleaseAllControllers)

	body, err := json.Marshal(map[string]any{
		"answers": []prompt.QuestionAnswer{{
			Question: "What scope?",
			Answer:   "API only",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/runs/"+runID+"/clarify", strings.NewReader(string(body)))
	req.SetPathValue("id", runID)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	api.ClarifyRun(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
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

type clarifyRestartRunner struct {
	workDir string
}

func (r *clarifyRestartRunner) Run(ctx context.Context, prompt string, _ chan<- runner.OutputLine) error {
	if strings.Contains(prompt, "implementation agent") {
		<-ctx.Done()
		return ctx.Err()
	}
	prdPath := filepath.Join(r.workDir, "prd.json")
	data := `{"version":1,"project_name":"Test","branch_name":"feature/x","stories":[{"id":"s1","title":"Story","description":"Do it","slices":[{"id":"slice-1","behavior":"AC","red_hint":"add failing test","passes":false}],"priority":1,"passes":false}]}`
	return os.WriteFile(prdPath, []byte(data), 0644)
}

func (r *clarifyRestartRunner) RunnerName() string        { return "mock" }
func (r *clarifyRestartRunner) CommandName() string       { return "mock" }
func (r *clarifyRestartRunner) IsInternalLog(string) bool { return false }
