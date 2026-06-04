package workflow

import (
	"context"
	"strings"
	"testing"

	"ralph/internal/shared/config"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
	"ralph/internal/workflow/review"
)

func TestRunImplementationReviewEventsAfterStoryCompleted(t *testing.T) {
	cfg, testPRD := saveSingleStoryPRDInGitRepo(t, false)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	exec := NewExecutorWithRunner(cfg, ch, mock)
	if err := exec.RunImplementation(context.Background(), testPRD); err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

	evts := drainEvents(ch)
	if err := assertReviewAfterStoryCompleted(evts); err != nil {
		t.Fatal(err)
	}
}

func assertReviewAfterStoryCompleted(evts []Event) error {
	seenStoryCompleted := false
	seenReviewStarted := false
	seenReviewCompleted := false
	for _, e := range evts {
		switch e.(type) {
		case EventStoryCompleted:
			seenStoryCompleted = true
		case EventImplementationReviewStarted:
			if !seenStoryCompleted {
				return errEventOrder{"implementation review started before story completed"}
			}
			seenReviewStarted = true
		case EventImplementationReviewCompleted:
			if !seenReviewStarted {
				return errEventOrder{"implementation review completed before started"}
			}
			seenReviewCompleted = true
		}
	}
	if !seenStoryCompleted {
		return errEventOrder{"missing EventStoryCompleted"}
	}
	if !seenReviewStarted {
		return errEventOrder{"missing EventImplementationReviewStarted"}
	}
	if !seenReviewCompleted {
		return errEventOrder{"missing EventImplementationReviewCompleted"}
	}
	return nil
}

type errEventOrder struct{ msg string }

func (e errEventOrder) Error() string { return e.msg }

func TestRunImplementationReviewBeforeNextStory(t *testing.T) {
	tmpDir := t.TempDir()
	initGitRepoInDir(t, tmpDir)
	cfg := config.DefaultConfig()
	cfg.WorkDir = tmpDir
	cfg.PRDFile = "prd.json"
	cfg.SkipCleanup = true

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories: []*prd.Story{
			{ID: "story-1", Title: "One", Description: "d", AcceptanceCriteria: []string{"a"}, Priority: 1, Passes: false},
			{ID: "story-2", Title: "Two", Description: "d", AcceptanceCriteria: []string{"a"}, Priority: 2, Passes: false},
		},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("save PRD: %v", err)
	}
	commitPRDFile(t, tmpDir, cfg.PRDFile)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	exec := NewExecutorWithRunner(cfg, ch, mock)
	if err := exec.RunImplementation(context.Background(), testPRD); err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

	if err := assertReviewBetweenStories(drainEvents(ch)); err != nil {
		t.Fatal(err)
	}
}

func assertReviewBetweenStories(evts []Event) error {
	var last string
	var storyStarts int
	for _, e := range evts {
		switch e.(type) {
		case EventStoryCompleted:
			last = "story_completed"
		case EventImplementationReviewStarted:
			if last != "story_completed" {
				return errEventOrder{"review started before story completed"}
			}
			last = "review_started"
		case EventImplementationReviewCompleted:
			if last != "review_started" {
				return errEventOrder{"review completed before started"}
			}
			last = "review_completed"
		case EventStoryStarted:
			storyStarts++
			if storyStarts > 1 && last != "review_completed" {
				return errEventOrder{"next story started before review completed"}
			}
			last = "story_started"
		}
	}
	if storyStarts < 2 {
		return errEventOrder{"expected two story starts"}
	}
	return nil
}

func TestRunImplementationReviewBeforeCleanup(t *testing.T) {
	cfg, testPRD := saveSingleStoryPRDInGitRepo(t, false)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	exec := NewExecutorWithRunner(cfg, ch, mock)
	if err := exec.RunImplementation(context.Background(), testPRD); err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

	if err := assertReviewBeforeCleanup(drainEvents(ch)); err != nil {
		t.Fatal(err)
	}
}

func assertReviewBeforeCleanup(evts []Event) error {
	lastStoryCompleted := -1
	reviewCompleted := -1
	cleanupStarted := -1
	for i, e := range evts {
		switch e.(type) {
		case EventStoryCompleted:
			lastStoryCompleted = i
		case EventImplementationReviewCompleted:
			reviewCompleted = i
		case EventCleanupStarted:
			cleanupStarted = i
		}
	}
	if lastStoryCompleted < 0 {
		return errEventOrder{"missing EventStoryCompleted"}
	}
	if reviewCompleted < 0 {
		return errEventOrder{"missing EventImplementationReviewCompleted"}
	}
	if cleanupStarted < 0 {
		return errEventOrder{"missing EventCleanupStarted"}
	}
	if reviewCompleted <= lastStoryCompleted {
		return errEventOrder{"review must follow last story completed"}
	}
	if cleanupStarted <= reviewCompleted {
		return errEventOrder{"cleanup must follow implementation review"}
	}
	return nil
}

func TestRunImplementationNoReviewBetweenStoryStartAndComplete(t *testing.T) {
	cfg, testPRD := saveSingleStoryPRDInGitRepo(t, true)

	ch := make(chan Event, 100)
	mock := newMockRunner()
	inStory := false
	mock.runFunc = func(_ context.Context, p string, _ chan<- runner.OutputLine) error {
		if strings.Contains(p, "critical diff review") && inStory {
			t.Fatal("implementation review runner invoked between story start and complete")
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	eventsDone := make(chan struct{})
	go func() {
		defer close(eventsDone)
		for ev := range ch {
			switch ev.(type) {
			case EventStoryStarted:
				inStory = true
			case EventStoryCompleted:
				inStory = false
			}
		}
	}()

	if err := exec.RunImplementation(context.Background(), testPRD); err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}
	close(ch)
	<-eventsDone
}

func TestRunImplementationDuplicateFingerprintSkipsThirdReviewRunner(t *testing.T) {
	workDir, _ := setupGitRepoWithWorkingTreeDiff(t)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"
	cfg.SkipCleanup = true

	testPRD := &prd.PRD{ProjectName: "Test", Context: "ctx"}
	findingsTranscript := `===ralph-findings===
[{"category":"bug","path":"delta.txt","summary":"missing test"}]
===/ralph-findings===`

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(_ context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if strings.Contains(p, "critical diff review") {
			outputCh <- runner.OutputLine{Text: findingsTranscript}
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	exec.runID = "run-dup"
	exec.reviewFingerprint = reviewFingerprintFromTranscript(t, findingsTranscript)

	if _, err := exec.runImplementationReview(context.Background(), testPRD); err == nil {
		t.Fatal("second review with duplicate fingerprint should return error")
	}

	storyCalls, reviewCalls, _ := countRunnerPromptKinds(mock)
	if storyCalls != 0 {
		t.Errorf("story runner calls = %d, want 0", storyCalls)
	}
	if reviewCalls != 1 {
		t.Errorf("review runner calls = %d, want 1 without a third review attempt", reviewCalls)
	}
}

func reviewFingerprintFromTranscript(t *testing.T, transcript string) string {
	t.Helper()
	findings, err := review.ParseFindings(transcript)
	if err != nil {
		t.Fatalf("ParseFindings() err = %v", err)
	}
	return review.Fingerprint(findings)
}

type recordingReviewLoop struct {
	iteration   int
	fingerprint string
	elapsedMs   int64
	stopReason  string
	updates     []ReviewLoopUpdate
}

func (r *recordingReviewLoop) Snapshot() (int, string, int64) {
	return r.iteration, r.fingerprint, r.elapsedMs
}

func (r *recordingReviewLoop) Apply(u ReviewLoopUpdate) error {
	r.iteration = u.ReviewIteration
	r.fingerprint = u.ReviewFingerprint
	r.elapsedMs = u.ReviewElapsedMs
	r.stopReason = u.StopReason
	r.updates = append(r.updates, u)
	return nil
}

func TestRunImplementationDuplicateFingerprintUpdatesStopReason(t *testing.T) {
	workDir, _ := setupGitRepoWithWorkingTreeDiff(t)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"

	findingsTranscript := `===ralph-findings===
[{"category":"bug","path":"delta.txt","summary":"missing test"}]
===/ralph-findings===`

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(_ context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if strings.Contains(p, "critical diff review") {
			outputCh <- runner.OutputLine{Text: findingsTranscript}
		}
		return nil
	}

	updater := &recordingReviewLoop{fingerprint: reviewFingerprintFromTranscript(t, findingsTranscript)}
	exec := NewExecutorWithRunner(cfg, ch, mock)
	exec.SetReviewLoop("run-stop", updater)

	_, _ = exec.runImplementationReview(context.Background(), &prd.PRD{Context: "ctx"})

	if updater.stopReason != StopReasonDuplicateFindings {
		t.Fatalf("stop_reason = %q, want %q", updater.stopReason, StopReasonDuplicateFindings)
	}
}

func TestRunImplementationFindingsEmitReviewAndStop(t *testing.T) {
	workDir, _ := setupGitRepoWithWorkingTreeDiff(t)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"
	cfg.SkipCleanup = true

	testPRD := &prd.PRD{
		ProjectName: "Test",
		Stories: []*prd.Story{
			{ID: "story-1", Title: "One", Description: "d", AcceptanceCriteria: []string{"a"}, Priority: 1, Passes: false},
			{ID: "story-2", Title: "Two", Description: "d", AcceptanceCriteria: []string{"a"}, Priority: 2, Passes: false},
		},
	}
	if err := prd.Save(cfg, testPRD); err != nil {
		t.Fatalf("save PRD: %v", err)
	}
	commitPRDFile(t, workDir, cfg.PRDFile)

	findingsTranscript := `===ralph-findings===
[{"category":"bug","path":"delta.txt","summary":"issue"}]
===/ralph-findings===`

	ch := make(chan Event, 100)
	mock := newMockRunner()
	mock.runFunc = func(_ context.Context, p string, outputCh chan<- runner.OutputLine) error {
		if strings.Contains(p, "critical diff review") {
			outputCh <- runner.OutputLine{Text: findingsTranscript}
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	if err := exec.RunImplementation(context.Background(), testPRD); err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

	evts := drainEvents(ch)
	var foundReview bool
	var foundSecondStory bool
	for _, e := range evts {
		if _, ok := e.(EventImplementationReview); ok {
			foundReview = true
		}
		if ev, ok := e.(EventStoryStarted); ok && ev.Story.ID == "story-2" {
			foundSecondStory = true
		}
	}
	if !foundReview {
		t.Fatal("expected EventImplementationReview with findings")
	}
	if foundSecondStory {
		t.Fatal("story-2 should not start after review findings")
	}
	storyCalls, reviewCalls, _ := countRunnerPromptKinds(mock)
	if storyCalls != 1 {
		t.Errorf("story runner calls = %d, want 1", storyCalls)
	}
	if reviewCalls != 1 {
		t.Errorf("review runner calls = %d, want 1", reviewCalls)
	}
}
