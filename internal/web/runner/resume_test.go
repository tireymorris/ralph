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
	"ralph/internal/shared/runstate"
	"ralph/internal/web/runs"
	"ralph/internal/workflow/events"
)

func cleanupController(t *testing.T, ctrl *RunController) {
	t.Helper()
	t.Cleanup(func() {
		ctrl.Cancel()
		time.Sleep(100 * time.Millisecond)
	})
}

func TestForceResumeRestartsExpectedPhaseForEachCheckpoint(t *testing.T) {
	tests := []struct {
		name        string
		checkpoint  string
		wantEvents  []string
		wantNoEvent bool
		wantStatus  string
		wantPhase   string
	}{
		{
			name:       "prd review reloads PRD for review",
			checkpoint: runs.CheckpointPRDReview,
			wantEvents: []string{"EventPRDLoaded", "EventPRDReview"},
			wantStatus: runstate.StatusWaitingReview,
			wantPhase:  runstate.PhaseReview,
		},
		{
			name:       "implementation review resumes implementation",
			checkpoint: runs.CheckpointImplReview,
			wantEvents: []string{"EventStoryStarted"},
			wantStatus: runstate.StatusImplementing,
			wantPhase:  runstate.PhaseImplement,
		},
		{
			name:       "followup resumes implementation",
			checkpoint: runs.CheckpointFollowup,
			wantEvents: []string{"EventStoryStarted"},
			wantStatus: runstate.StatusImplementing,
			wantPhase:  runstate.PhaseImplement,
		},
		{
			name:        "complete starts no phase",
			checkpoint:  runs.CheckpointComplete,
			wantNoEvent: true,
			wantStatus:  runstate.StatusCompleted,
			wantPhase:   runstate.PhaseCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workDir := t.TempDir()
			reg := runs.NewRegistry()
			run := &runs.Run{
				ID:         "run-" + strings.ReplaceAll(tt.checkpoint, "_", "-"),
				WorkDir:    workDir,
				Prompt:     "build feature",
				Status:     "running",
				Phase:      "resume",
				Checkpoint: tt.checkpoint,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				PRDPath:    "prd.json",
			}
			if err := reg.Register(run); err != nil {
				t.Fatalf("Register() error = %v", err)
			}
			writeForceResumePRD(t, workDir)

			got, ok := reg.Get(run.ID)
			if !ok {
				t.Fatal("registered run missing")
			}
			if got.Checkpoint != tt.checkpoint {
				t.Fatalf("checkpoint = %q, want %q", got.Checkpoint, tt.checkpoint)
			}

			cfg := config.DefaultConfig()
			cfg.WorkDir = workDir
			cfg.PRDFile = "prd.json"
			cfg.SkipCleanup = true

			ctrl := NewControllerWithRunner(cfg, reg, run.ID, &testRunner{})
			cleanupController(t, ctrl)
			ch, unsub := ctrl.Subscribe()
			defer unsub()

			ctrl.ForceResume(context.Background())

			if tt.wantNoEvent {
				assertNoForceResumeEvent(t, ch)
			} else {
				for _, want := range tt.wantEvents {
					ev := nextForceResumeEvent(t, ch)
					if forceResumeEventName(ev) != want {
						t.Fatalf("event = %s, want %s", forceResumeEventName(ev), want)
					}
				}
			}
			deadline := time.Now().Add(2 * time.Second)
			for time.Now().Before(deadline) {
				got, ok := reg.Get(run.ID)
				if ok && got.Status == tt.wantStatus && got.Phase == tt.wantPhase {
					return
				}
				time.Sleep(10 * time.Millisecond)
			}
			state, _ := reg.Get(run.ID)
			t.Fatalf("run state = %q/%q, want %q/%q", state.Status, state.Phase, tt.wantStatus, tt.wantPhase)
		})
	}
}

func writeForceResumePRD(t *testing.T, workDir string) {
	t.Helper()
	data := `{"version":1,"project_name":"Test","branch_name":"feature/x","stories":[{"id":"s1","title":"Story","description":"Do it","acceptance_criteria":["AC"],"priority":1,"passes":false}]}`
	if err := os.WriteFile(filepath.Join(workDir, "prd.json"), []byte(data), 0o644); err != nil {
		t.Fatalf("WriteFile prd: %v", err)
	}
}

func nextForceResumeEvent(t *testing.T, ch <-chan events.Event) events.Event {
	t.Helper()
	select {
	case ev := <-ch:
		return ev
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for ForceResume event")
	}
	return nil
}

func assertNoForceResumeEvent(t *testing.T, ch <-chan events.Event) {
	t.Helper()
	select {
	case ev := <-ch:
		t.Fatalf("unexpected event after complete checkpoint resume: %s", forceResumeEventName(ev))
	case <-time.After(300 * time.Millisecond):
	}
}

func forceResumeEventName(ev events.Event) string {
	switch ev.(type) {
	case events.EventPRDLoaded:
		return "EventPRDLoaded"
	case events.EventPRDReview:
		return "EventPRDReview"
	case events.EventStoryStarted:
		return "EventStoryStarted"
	case events.EventCompleted:
		return "EventCompleted"
	case events.EventError:
		return "EventError"
	default:
		return "unknown"
	}
}

func TestForceResumeImplReviewCheckpointOverridesWaitingReviewStatus(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:         "run-impl-stale-status",
		WorkDir:    workDir,
		Prompt:     "build feature",
		Status:     "waiting_review",
		Phase:      "review",
		Checkpoint: runs.CheckpointImplReview,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		PRDPath:    "prd.json",
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	prdPath := filepath.Join(workDir, "prd.json")
	data := `{"version":1,"project_name":"Test","branch_name":"feature/x","stories":[{"id":"s1","title":"Story","description":"Do it","acceptance_criteria":["AC"],"priority":1,"passes":false}]}`
	if err := os.WriteFile(prdPath, []byte(data), 0644); err != nil {
		t.Fatalf("WriteFile prd: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	mock := &testRunner{
		runFunc: func(ctx context.Context, prompt string, _ chan<- runner.OutputLine) error {
			if strings.Contains(prompt, "You are Ralph's implementation agent") {
				<-ctx.Done()
				return ctx.Err()
			}
			return nil
		},
	}
	ctrl := NewControllerWithRunner(cfg, reg, run.ID, mock)
	t.Cleanup(func() {
		ctrl.Cancel()
		time.Sleep(100 * time.Millisecond)
	})

	ch, unsub := ctrl.Subscribe()
	defer unsub()

	ctrl.ForceResume(context.Background())

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case ev := <-ch:
			switch ev.(type) {
			case events.EventPRDReview:
				t.Fatal("ForceResume emitted EventPRDReview when checkpoint is impl_review")
			case events.EventPRDLoaded:
				t.Fatal("ForceResume emitted EventPRDLoaded when checkpoint is impl_review")
			default:
			}
		default:
		}
		got, ok := reg.Get(run.ID)
		if ok && strings.Contains(got.Phase, "implement") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("run did not resume implementation from impl_review checkpoint")
}

func TestForceResumeImplReviewCheckpointSkipsPRDGenerating(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:         "run-impl-ckpt",
		WorkDir:    workDir,
		Prompt:     "build feature",
		Status:     "implementing",
		Phase:      "implement",
		Checkpoint: runs.CheckpointImplReview,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		PRDPath:    "prd.json",
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	prdPath := filepath.Join(workDir, "prd.json")
	data := `{"version":1,"project_name":"Test","branch_name":"feature/x","stories":[{"id":"s1","title":"Story","description":"Do it","acceptance_criteria":["AC"],"priority":1,"passes":false}]}`
	if err := os.WriteFile(prdPath, []byte(data), 0644); err != nil {
		t.Fatalf("WriteFile prd: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	mock := &testRunner{
		runFunc: func(ctx context.Context, prompt string, _ chan<- runner.OutputLine) error {
			if strings.Contains(prompt, "You are Ralph's implementation agent") {
				<-ctx.Done()
				return ctx.Err()
			}
			return nil
		},
	}
	ctrl := NewControllerWithRunner(cfg, reg, run.ID, mock)
	t.Cleanup(func() {
		ctrl.Cancel()
		time.Sleep(100 * time.Millisecond)
	})

	ch, unsub := ctrl.Subscribe()
	defer unsub()

	ctrl.ForceResume(context.Background())

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case ev := <-ch:
			switch ev.(type) {
			case events.EventPRDGenerating:
				t.Fatal("ForceResume emitted EventPRDGenerating with impl_review checkpoint")
			case events.EventPRDLoaded:
				t.Fatal("ForceResume emitted EventPRDLoaded with impl_review checkpoint")
			default:
			}
		default:
		}
		got, ok := reg.Get(run.ID)
		if ok && strings.Contains(got.Phase, "implement") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("run did not return to implement phase after ForceResume with impl_review checkpoint")
}

func TestForceResumeContinuesImplementation(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:        "run-resume",
		WorkDir:   workDir,
		Prompt:    "build feature",
		Status:    "implementing",
		Phase:     "implement",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		PRDPath:   "prd.json",
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	prdPath := filepath.Join(workDir, "prd.json")
	data := `{"version":1,"project_name":"Test","branch_name":"feature/x","stories":[{"id":"s1","title":"Story","description":"Do it","acceptance_criteria":["AC"],"priority":1,"passes":false}]}`
	if err := os.WriteFile(prdPath, []byte(data), 0644); err != nil {
		t.Fatalf("WriteFile prd: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	mock := &testRunner{
		runFunc: func(ctx context.Context, prompt string, _ chan<- runner.OutputLine) error {
			if strings.Contains(prompt, "You are Ralph's implementation agent") {
				<-ctx.Done()
				return ctx.Err()
			}
			return nil
		},
	}
	ctrl := NewControllerWithRunner(cfg, reg, run.ID, mock)
	t.Cleanup(func() {
		ctrl.Cancel()
		time.Sleep(100 * time.Millisecond)
	})

	ctrl.ForceResume(context.Background())

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		got, ok := reg.Get(run.ID)
		if ok && strings.Contains(got.Phase, "implement") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("run did not return to implement phase after ForceResume")
}

func TestForceResumeReloadsPRDForWaitingReview(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:        "run-review",
		WorkDir:   workDir,
		Prompt:    "build feature",
		Status:    "waiting_review",
		Phase:     "review",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		PRDPath:   "prd.json",
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	prdPath := filepath.Join(workDir, "prd.json")
	data := `{"version":1,"project_name":"Test","branch_name":"feature/x","stories":[{"id":"s1","title":"Story","description":"Do it","acceptance_criteria":["AC"],"priority":1,"passes":false}]}`
	if err := os.WriteFile(prdPath, []byte(data), 0644); err != nil {
		t.Fatalf("WriteFile prd: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	ctrl := NewControllerWithRunner(cfg, reg, run.ID, &testRunner{})
	t.Cleanup(func() {
		ctrl.Cancel()
		time.Sleep(100 * time.Millisecond)
	})

	ctrl.ForceResume(context.Background())

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		got, ok := reg.Get(run.ID)
		if ok && got.Status == "waiting_review" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("run did not return to waiting_review after ForceResume")
}

func TestForceResumeRestartsFromPromptWithoutPRD(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:        "run-restart",
		WorkDir:   workDir,
		Prompt:    "build feature",
		Status:    "running",
		Phase:     "clarify",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		PRDPath:   "prd.json",
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	ctrl := NewControllerWithRunner(cfg, reg, run.ID, &testRunner{})
	t.Cleanup(func() {
		ctrl.Cancel()
		time.Sleep(100 * time.Millisecond)
	})

	ctrl.ForceResume(context.Background())

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		got, ok := reg.Get(run.ID)
		if ok && got.Phase == "clarify" && got.Status == "running" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("run did not restart clarify phase after ForceResume without PRD")
}

func TestForceResumeIgnoresCachedPRDWhenDiskPRDIsMissing(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:        "run-stale-cache",
		WorkDir:   workDir,
		Prompt:    "",
		Status:    "running",
		Phase:     "implement",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		PRDPath:   "prd.json",
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	ctrl := NewControllerWithRunner(cfg, reg, run.ID, &testRunner{})
	cleanupController(t, ctrl)
	ctrl.TrackEventState(events.EventPRDGenerated{PRD: testPRD("cached")})

	ch, unsub := ctrl.Subscribe()
	defer unsub()

	ctrl.ForceResume(context.Background())

	select {
	case ev := <-ch:
		errEv, ok := ev.(events.EventError)
		if !ok {
			t.Fatalf("event = %T, want EventError", ev)
		}
		if errEv.Err == nil || !strings.Contains(errEv.Err.Error(), "cannot resume") {
			t.Fatalf("error = %v, want cannot resume", errEv.Err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for EventError")
	}
}

func testPRD(name string) *prd.PRD {
	return &prd.PRD{
		Version:     1,
		ProjectName: name,
		Stories: []*prd.Story{{
			ID:                 "s1",
			Title:              "Story",
			Description:        "Do it",
			AcceptanceCriteria: []string{"AC"},
			Priority:           1,
		}},
	}
}

func TestForceResumeEmitsErrorWithoutPRDOrPrompt(t *testing.T) {
	workDir := t.TempDir()
	reg := runs.NewRegistry()
	run := &runs.Run{
		ID:        "run-empty",
		WorkDir:   workDir,
		Prompt:    "",
		Status:    "running",
		Phase:     "clarify",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		PRDPath:   "prd.json",
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	ctrl := NewControllerWithRunner(cfg, reg, run.ID, &testRunner{})
	cleanupController(t, ctrl)

	ch, unsub := ctrl.Subscribe()
	defer unsub()

	ctrl.ForceResume(context.Background())

	select {
	case ev := <-ch:
		errEv, ok := ev.(events.EventError)
		if !ok {
			t.Fatalf("event = %T, want EventError", ev)
		}
		if errEv.Err == nil || !strings.Contains(errEv.Err.Error(), "cannot resume") {
			t.Fatalf("error = %v, want cannot resume", errEv.Err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for EventError")
	}
}
