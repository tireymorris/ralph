package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
	"ralph/internal/web/runs"
	"ralph/internal/workflow/events"
)

func TestFollowUpResumeTransitionsToImplementPhase(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:        "run-followup",
		WorkDir:   workDir,
		Prompt:    "build feature",
		Status:    "completed",
		Phase:     "complete",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		PRDPath:   "prd.json",
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	prdPath := filepath.Join(workDir, "prd.json")
	data := `{"version":1,"project_name":"Test","branch_name":"feature/x","stories":[{"id":"s1","title":"Story","description":"Do it","acceptance_criteria":["AC"],"priority":1}]}`
	if err := os.WriteFile(prdPath, []byte(data), 0644); err != nil {
		t.Fatalf("WriteFile prd: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	mock := &testRunner{
		runFunc: func(ctx context.Context, prompt string, _ chan<- runner.OutputLine) error {
			if strings.Contains(prompt, "implementation agent") {
				<-ctx.Done()
				return ctx.Err()
			}
			return nil
		},
	}
	ctrl := NewControllerWithRunner(cfg, reg, run.ID, mock)
	t.Cleanup(ctrl.Cancel)

	p, err := prd.Load(cfg)
	if err != nil {
		t.Fatalf("prd.Load: %v", err)
	}
	p.UnmarkAllStories()
	if err := prd.Save(cfg, p); err != nil {
		t.Fatalf("prd.Save: %v", err)
	}
	p, err = prd.Load(cfg)
	if err != nil {
		t.Fatalf("prd.Load after save: %v", err)
	}

	_ = reg.UpdateStatus(run.ID, "running", "followup")
	ctx := context.Background()
	ctrl.StartImplementation(ctx, p)

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		got, ok := reg.Get(run.ID)
		if ok && strings.Contains(got.Phase, "implement") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	got, _ := reg.Get(run.ID)
	t.Fatalf("registry phase = %q status = %q, want phase containing implement within 5s", got.Phase, got.Status)
}

func TestFollowUpForwardsRunnerOutput(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:      "run-followup-output",
		WorkDir: workDir,
		Prompt:  "build feature",
		Status:  "completed",
		Phase:   "complete",
		PRDPath: "prd.json",
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	prdPath := filepath.Join(workDir, "prd.json")
	prdData := `{"version":1,"project_name":"Test","stories":[{"id":"s1","title":"Story","description":"Do it","acceptance_criteria":["AC"],"priority":1}]}`
	if err := os.WriteFile(prdPath, []byte(prdData), 0644); err != nil {
		t.Fatal(err)
	}

	eventsDir := filepath.Join(workDir, ".ralph", "runs", run.ID)
	if err := os.MkdirAll(eventsDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(eventsDir, "events.ndjson"), []byte(""), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	mock := &testRunner{
		runFunc: func(ctx context.Context, prompt string, outputCh chan<- runner.OutputLine) error {
			if strings.Contains(prompt, "follow-up") || strings.Contains(prompt, "revision") {
				outputCh <- runner.OutputLine{Text: "revising PRD..."}
				outputCh <- runner.OutputLine{Text: "done revising"}
				return nil
			}
			<-ctx.Done()
			return ctx.Err()
		},
	}
	ctrl := NewControllerWithRunner(cfg, reg, run.ID, mock)
	t.Cleanup(ctrl.Cancel)

	sub, unsub := ctrl.Subscribe()
	defer unsub()

	go ctrl.RunFollowUp(context.Background(), "add logging", cfg)

	deadline := time.Now().Add(5 * time.Second)
	var outputTexts []string
	for time.Now().Before(deadline) {
		select {
		case ev := <-sub:
			if o, ok := ev.(events.EventOutput); ok {
				outputTexts = append(outputTexts, o.Text)
			}
		default:
		}
		if len(outputTexts) >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if len(outputTexts) < 2 {
		t.Fatalf("expected at least 2 output events, got %d: %v", len(outputTexts), outputTexts)
	}
	found := false
	for _, txt := range outputTexts {
		if strings.Contains(txt, "revising") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected output containing 'revising', got %v", outputTexts)
	}
}
