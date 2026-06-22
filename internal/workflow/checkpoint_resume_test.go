package workflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runstate"
	"ralph/internal/workflow/events"
)

func TestStartCheckpointResumeRestartsExpectedPhase(t *testing.T) {
	tests := []struct {
		name        string
		checkpoint  string
		wantEvents  []string
		wantNoEvent bool
	}{
		{
			name:       "prd review reloads PRD for review",
			checkpoint: runstate.CheckpointPRDReview,
			wantEvents: []string{"EventPRDLoaded", "EventPRDReview"},
		},
		{
			name:       "implementation review resumes implementation",
			checkpoint: runstate.CheckpointImplReview,
			wantEvents: []string{"EventStoryStarted"},
		},
		{
			name:       "followup resumes implementation",
			checkpoint: runstate.CheckpointFollowup,
			wantEvents: []string{"EventStoryStarted"},
		},
		{
			name:        "complete starts no phase",
			checkpoint:  runstate.CheckpointComplete,
			wantNoEvent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workDir := t.TempDir()
			cfg := config.DefaultConfig()
			cfg.WorkDir = workDir
			cfg.PRDFile = "prd.json"
			cfg.SkipCleanup = true

			writeCheckpointResumePRD(t, workDir)
			loop := NewFileReviewLoop(workDir, runstate.LocalRunID)
			if err := loop.Apply(ReviewLoopUpdate{Checkpoint: tt.checkpoint}); err != nil {
				t.Fatalf("seed checkpoint: %v", err)
			}
			if got := loop.Checkpoint(); got != tt.checkpoint {
				t.Fatalf("checkpoint = %q, want %q", got, tt.checkpoint)
			}

			d := NewDriverWithRunner(cfg, newMockRunner())
			t.Cleanup(d.Cancel)
			d.SetReviewLoop(runstate.LocalRunID, loop)
			d.StartCheckpointResume(context.Background())

			if tt.wantNoEvent {
				assertNoCheckpointResumeEvent(t, d.EventsCh())
				return
			}

			for _, want := range tt.wantEvents {
				got := nextCheckpointResumeEvent(t, d.EventsCh())
				if eventName(got) != want {
					t.Fatalf("event = %s, want %s", eventName(got), want)
				}
			}
		})
	}
}

func TestStartCheckpointResumeUsesInjectedStoreForImplementationResume(t *testing.T) {
	workDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"
	cfg.SkipCleanup = true

	seeded := &prd.PRD{
		ProjectName: "Injected",
		Stories: []*prd.Story{
			{
				ID:                 "story-1",
				Title:              "Story",
				Description:        "Desc",
				Slices:      testStorySlice("AC"),
				Priority:           1,
			},
		},
	}

	loop := NewFileReviewLoop(workDir, runstate.LocalRunID)
	if err := loop.Apply(ReviewLoopUpdate{Checkpoint: runstate.CheckpointImplReview}); err != nil {
		t.Fatalf("seed checkpoint: %v", err)
	}

	d := NewDriverWithRunner(cfg, newMockRunner())
	t.Cleanup(d.Cancel)
	d.executor = NewExecutorWithRunnerAndStore(cfg, d.eventsCh, newMockRunner(), inMemoryPRDStore{p: seeded})
	d.SetReviewLoop(runstate.LocalRunID, loop)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d.StartCheckpointResume(ctx)

	got := nextCheckpointResumeEvent(t, d.EventsCh())
	if eventName(got) != "EventStoryStarted" {
		t.Fatalf("event = %s, want EventStoryStarted", eventName(got))
	}
	if _, err := os.Stat(filepath.Join(workDir, cfg.PRDFile)); !os.IsNotExist(err) {
		t.Fatalf("expected no prd.json on disk, stat error = %v", err)
	}
	cancel()
	d.Wait()
}

func writeCheckpointResumePRD(t *testing.T, workDir string) {
	t.Helper()
	data := `{"project_name":"Checkpoint","stories":[{"id":"story-1","title":"Story 1","description":"d","slices":[{"id":"slice-1","behavior":"a","red_hint":"add failing test"}],"priority":1}]}`
	if err := os.WriteFile(filepath.Join(workDir, "prd.json"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func nextCheckpointResumeEvent(t *testing.T, ch <-chan events.Event) events.Event {
	t.Helper()
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	for {
		select {
		case ev := <-ch:
			return ev
		case <-timer.C:
			t.Fatal("timed out waiting for checkpoint resume event")
		}
	}
}

func assertNoCheckpointResumeEvent(t *testing.T, ch <-chan events.Event) {
	t.Helper()
	select {
	case ev := <-ch:
		t.Fatalf("unexpected event after complete checkpoint resume: %s", eventName(ev))
	case <-time.After(300 * time.Millisecond):
	}
}

func eventName(ev events.Event) string {
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
