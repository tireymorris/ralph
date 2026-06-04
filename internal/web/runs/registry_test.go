package runs

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRegisterGet(t *testing.T) {
	reg := NewRegistry()
	workDir := t.TempDir()

	run := &Run{
		ID:        "run-1",
		WorkDir:   workDir,
		Prompt:    "build feature",
		Status:    "running",
		Phase:     "clarify",
		CreatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		PRDPath:   "prd.json",
	}

	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	got, ok := reg.Get("run-1")
	if !ok {
		t.Fatal("Get() ok = false, want true")
	}
	if got.Prompt != run.Prompt {
		t.Errorf("Prompt = %q, want %q", got.Prompt, run.Prompt)
	}
}

func TestListSortedByCreatedAtDesc(t *testing.T) {
	reg := NewRegistry()
	workDir := t.TempDir()

	oldest := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	middle := time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC)
	newest := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	for _, run := range []*Run{
		{ID: "run-old", WorkDir: workDir, CreatedAt: oldest, UpdatedAt: oldest},
		{ID: "run-mid", WorkDir: workDir, CreatedAt: middle, UpdatedAt: middle},
		{ID: "run-new", WorkDir: workDir, CreatedAt: newest, UpdatedAt: newest},
	} {
		if err := reg.Register(run); err != nil {
			t.Fatalf("Register(%s) error = %v", run.ID, err)
		}
	}

	list := reg.List()
	if len(list) != 3 {
		t.Fatalf("len(List()) = %d, want 3", len(list))
	}
	want := []string{"run-new", "run-mid", "run-old"}
	for i, id := range want {
		if list[i].ID != id {
			t.Errorf("List()[%d].ID = %q, want %q", i, list[i].ID, id)
		}
	}
}

func TestUpdateStatus(t *testing.T) {
	reg := NewRegistry()
	workDir := t.TempDir()

	run := &Run{
		ID:        "run-upd",
		WorkDir:   workDir,
		Status:    "running",
		Phase:     "clarify",
		CreatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if err := reg.UpdateStatus("run-upd", "completed", "done"); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	got, ok := reg.Get("run-upd")
	if !ok {
		t.Fatal("Get() ok = false")
	}
	if got.Status != "completed" || got.Phase != "done" {
		t.Errorf("status/phase = %q/%q, want completed/done", got.Status, got.Phase)
	}
	if !got.UpdatedAt.After(run.UpdatedAt) {
		t.Errorf("UpdatedAt = %v, want after %v", got.UpdatedAt, run.UpdatedAt)
	}
}

func TestUpdateCheckpoint(t *testing.T) {
	reg := NewRegistry()
	workDir := t.TempDir()

	run := &Run{
		ID:         "run-ckpt",
		WorkDir:    workDir,
		Status:     "waiting_review",
		Phase:      "review",
		Checkpoint: CheckpointPRDReview,
		CreatedAt:  time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if err := reg.UpdateCheckpoint("run-ckpt", CheckpointImplReview); err != nil {
		t.Fatalf("UpdateCheckpoint() error = %v", err)
	}

	got, ok := reg.Get("run-ckpt")
	if !ok {
		t.Fatal("Get() ok = false")
	}
	if got.Checkpoint != CheckpointImplReview {
		t.Errorf("Checkpoint = %q, want %q", got.Checkpoint, CheckpointImplReview)
	}
	if !got.UpdatedAt.After(run.UpdatedAt) {
		t.Errorf("UpdatedAt = %v, want after %v", got.UpdatedAt, run.UpdatedAt)
	}

	reloaded := NewRegistry()
	if err := reloaded.LoadFromWorkDir(workDir); err != nil {
		t.Fatalf("LoadFromWorkDir() error = %v", err)
	}
	got, ok = reloaded.Get("run-ckpt")
	if !ok {
		t.Fatal("Get() ok = false after reload")
	}
	if got.Checkpoint != CheckpointImplReview {
		t.Errorf("reloaded Checkpoint = %q, want %q", got.Checkpoint, CheckpointImplReview)
	}
}

func TestReviewLoopFieldsRoundTrip(t *testing.T) {
	workDir := t.TempDir()
	reg := NewRegistry()

	run := &Run{
		ID:                       "run-review-loop",
		WorkDir:                  workDir,
		Prompt:                   "build feature",
		Status:                   "implementing",
		Phase:                    "implement",
		CreatedAt:                time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC),
		UpdatedAt:                time.Date(2026, 6, 4, 12, 30, 0, 0, time.UTC),
		PRDPath:                  "prd.json",
		Checkpoint:               CheckpointImplReview,
		ReviewIteration:          2,
		ReviewFingerprint:        "abc123def4567890abc123def4567890abc123def4567890abc123def4567890",
		ReviewElapsedMs:          1500,
		StopReason:               "duplicate_findings",
		LastReviewTranscriptPath: "review-2.txt",
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	reloaded := NewRegistry()
	if err := reloaded.LoadFromWorkDir(workDir); err != nil {
		t.Fatalf("LoadFromWorkDir() error = %v", err)
	}

	got, ok := reloaded.Get("run-review-loop")
	if !ok {
		t.Fatal("Get() ok = false after reload")
	}
	if got.Checkpoint != run.Checkpoint {
		t.Errorf("Checkpoint = %q, want %q", got.Checkpoint, run.Checkpoint)
	}
	if got.ReviewIteration != run.ReviewIteration {
		t.Errorf("ReviewIteration = %d, want %d", got.ReviewIteration, run.ReviewIteration)
	}
	if got.ReviewFingerprint != run.ReviewFingerprint {
		t.Errorf("ReviewFingerprint = %q, want %q", got.ReviewFingerprint, run.ReviewFingerprint)
	}
	if got.ReviewElapsedMs != run.ReviewElapsedMs {
		t.Errorf("ReviewElapsedMs = %d, want %d", got.ReviewElapsedMs, run.ReviewElapsedMs)
	}
	if got.StopReason != run.StopReason {
		t.Errorf("StopReason = %q, want %q", got.StopReason, run.StopReason)
	}
	if got.LastReviewTranscriptPath != run.LastReviewTranscriptPath {
		t.Errorf("LastReviewTranscriptPath = %q, want %q", got.LastReviewTranscriptPath, run.LastReviewTranscriptPath)
	}
}

func TestUpdateReviewLoop(t *testing.T) {
	reg := NewRegistry()
	workDir := t.TempDir()

	run := &Run{
		ID:        "run-review-upd",
		WorkDir:   workDir,
		Status:    "implementing",
		Phase:     "implement",
		CreatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	fp := "abc123def4567890abc123def4567890abc123def4567890abc123def4567890"
	if err := reg.UpdateReviewLoop("run-review-upd", ReviewLoopUpdate{
		Checkpoint:               CheckpointImplReview,
		ReviewIteration:          3,
		ReviewFingerprint:        fp,
		ReviewElapsedMs:          4200,
		StopReason:               "guardrail",
		LastReviewTranscriptPath: "review-3.txt",
	}); err != nil {
		t.Fatalf("UpdateReviewLoop() error = %v", err)
	}

	got, ok := reg.Get("run-review-upd")
	if !ok {
		t.Fatal("Get() ok = false")
	}
	if got.Checkpoint != CheckpointImplReview {
		t.Errorf("Checkpoint = %q, want %q", got.Checkpoint, CheckpointImplReview)
	}
	if got.ReviewIteration != 3 {
		t.Errorf("ReviewIteration = %d, want 3", got.ReviewIteration)
	}
	if got.ReviewFingerprint != fp {
		t.Errorf("ReviewFingerprint = %q, want %q", got.ReviewFingerprint, fp)
	}
	if got.ReviewElapsedMs != 4200 {
		t.Errorf("ReviewElapsedMs = %d, want 4200", got.ReviewElapsedMs)
	}
	if got.StopReason != "guardrail" {
		t.Errorf("StopReason = %q, want guardrail", got.StopReason)
	}
	if got.LastReviewTranscriptPath != "review-3.txt" {
		t.Errorf("LastReviewTranscriptPath = %q, want review-3.txt", got.LastReviewTranscriptPath)
	}

	reloaded := NewRegistry()
	if err := reloaded.LoadFromWorkDir(workDir); err != nil {
		t.Fatalf("LoadFromWorkDir() error = %v", err)
	}
	got, ok = reloaded.Get("run-review-upd")
	if !ok {
		t.Fatal("Get() ok = false after reload")
	}
	if got.ReviewFingerprint != fp {
		t.Errorf("reloaded ReviewFingerprint = %q, want %q", got.ReviewFingerprint, fp)
	}
}

func TestResumeCheckpointFromReloadedMeta(t *testing.T) {
	workDir := t.TempDir()
	reg := NewRegistry()

	run := &Run{
		ID:         "run-resume-ckpt",
		WorkDir:    workDir,
		Prompt:     "build feature",
		Status:     "implementing",
		Phase:      "implement",
		Checkpoint: CheckpointImplReview,
		CreatedAt:  time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		PRDPath:    "prd.json",
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	reloaded := NewRegistry()
	if err := reloaded.LoadFromWorkDir(workDir); err != nil {
		t.Fatalf("LoadFromWorkDir() error = %v", err)
	}

	got, ok := reloaded.Get("run-resume-ckpt")
	if !ok {
		t.Fatal("Get() ok = false after reload")
	}
	if got.Checkpoint != CheckpointImplReview {
		t.Errorf("Checkpoint = %q, want %q for resume", got.Checkpoint, CheckpointImplReview)
	}
}

func TestLoadFromWorkDir(t *testing.T) {
	workDir := t.TempDir()
	reg := NewRegistry()

	run := &Run{
		ID:        "run-persisted",
		WorkDir:   workDir,
		Prompt:    "saved run",
		Status:    "completed",
		Phase:     "complete",
		CreatedAt: time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC),
		PRDPath:   "prd.json",
	}
	if err := reg.Register(run); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	reloaded := NewRegistry()
	if err := reloaded.LoadFromWorkDir(workDir); err != nil {
		t.Fatalf("LoadFromWorkDir() error = %v", err)
	}

	got, ok := reloaded.Get("run-persisted")
	if !ok {
		t.Fatal("Get() ok = false after reload")
	}
	if got.Prompt != run.Prompt || got.Status != run.Status {
		t.Errorf("reloaded run = %+v, want prompt/status from registered run", got)
	}
}

func TestRegistryClear(t *testing.T) {
	reg := NewRegistry()
	reg.Clear()

	workDir := t.TempDir()
	ts := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	for _, id := range []string{"run-a", "run-b"} {
		run := &Run{
			ID:        id,
			WorkDir:   workDir,
			CreatedAt: ts,
			UpdatedAt: ts,
		}
		if err := reg.Register(run); err != nil {
			t.Fatalf("Register(%s) error = %v", id, err)
		}
	}

	reg.Clear()

	if len(reg.List()) != 0 {
		t.Fatalf("len(List()) = %d, want 0 after Clear", len(reg.List()))
	}
	if _, ok := reg.Get("run-a"); ok {
		t.Fatal("Get(run-a) ok = true after Clear, want false")
	}
}

func TestConcurrentRegisterList(t *testing.T) {
	reg := NewRegistry()
	workDir := t.TempDir()

	const n = 10
	var wg sync.WaitGroup
	wg.Add(n)
	for i := range n {
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("run-%d", i)
			run := &Run{
				ID:        id,
				WorkDir:   workDir,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
			if err := reg.Register(run); err != nil {
				t.Errorf("Register(%s) error = %v", id, err)
			}
		}(i)
	}

	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				_ = reg.List()
			}
		}
	}()

	wg.Wait()
	close(done)

	if len(reg.List()) != n {
		t.Fatalf("len(List()) = %d, want %d", len(reg.List()), n)
	}
}
