package workflow

import (
	"context"
	"strings"
	"testing"
	"time"

	"ralph/internal/prompt"
	"ralph/internal/shared/config"
	"ralph/internal/shared/gitdiff"
	"ralph/internal/shared/prd"
	"ralph/internal/shared/runner"
	"ralph/internal/shared/runstate"
	"ralph/internal/workflow/review"
)

func TestRunImplementationReviewEventsAfterStoryCompleted(t *testing.T) {
	cfg, testPRD := saveSingleStoryPRD(t, false)

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
	cfg, testPRD := saveSingleStoryPRD(t, false)

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
	cfg, testPRD := saveSingleStoryPRD(t, true)

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

func TestRunImplementationReviewOnceTimesOutReviewRunner(t *testing.T) {
	workDir, _ := setupGitRepoWithWorkingTreeDiff(t)
	cfg := config.DefaultConfig()
	cfg.WorkDir = workDir
	cfg.PRDFile = "prd.json"
	cfg.RunnerTimeout = 50 * time.Millisecond

	mock := newMockRunner()
	mock.runFunc = func(ctx context.Context, p string, _ chan<- runner.OutputLine) error {
		if strings.Contains(p, "critical diff review") {
			<-ctx.Done()
			return ctx.Err()
		}
		return nil
	}
	exec := NewExecutorWithRunner(cfg, make(chan Event, 10), mock)

	start := time.Now()
	_, err := exec.runImplementationReviewOnce(context.Background(), &prd.PRD{ProjectName: "Test"})
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("runImplementationReviewOnce() error = nil, want timeout error")
	}
	if !strings.Contains(err.Error(), "50ms") {
		t.Errorf("runImplementationReviewOnce() error = %v, want mention 50ms", err)
	}
	if elapsed > time.Second {
		t.Errorf("runImplementationReviewOnce() elapsed = %v, want under 1s", elapsed)
	}
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
	changed, err := gitdiff.ChangedFiles(workDir)
	if err != nil {
		t.Fatalf("ChangedFiles() err = %v", err)
	}
	exec.reviewFingerprint = reviewFingerprintFromTranscript(t, findingsTranscript)
	exec.reviewChangedFilesHash = gitdiff.HashFiles(changed)

	_, err = exec.runImplementationReview(context.Background(), testPRD)
	if err == nil {
		t.Fatal("duplicate fingerprint should return error after recovery is exhausted")
	}
	if !strings.Contains(err.Error(), runstate.StopReasonRecoveryExhausted) {
		t.Fatalf("error = %v, want %s", err, runstate.StopReasonRecoveryExhausted)
	}

	storyCalls, reviewCalls, _ := countRunnerPromptKinds(mock)
	if storyCalls < 2 {
		t.Errorf("recovery runner calls = %d, want at least 2", storyCalls)
	}
	if reviewCalls != 0 {
		t.Errorf("review runner calls = %d, want 0 when duplicate diff and fingerprint", reviewCalls)
	}
}

func reviewFingerprintFromTranscript(t *testing.T, transcript string) string {
	t.Helper()
	findings, err := review.ParseFindings(transcript, true)
	if err != nil {
		t.Fatalf("ParseFindings() err = %v", err)
	}
	return review.Fingerprint(findings)
}

type recordingReviewLoop struct {
	iteration        int
	fingerprint      string
	elapsedMs        int64
	changedFilesHash string
	stopReason       string
	updates          []ReviewLoopUpdate
}

func (r *recordingReviewLoop) Snapshot() (int, string, int64, string) {
	return r.iteration, r.fingerprint, r.elapsedMs, r.changedFilesHash
}

func (r *recordingReviewLoop) Apply(u ReviewLoopUpdate) error {
	r.iteration = u.ReviewIteration
	r.fingerprint = u.ReviewFingerprint
	r.elapsedMs = u.ReviewElapsedMs
	if u.LastReviewChangedFilesHash != "" {
		r.changedFilesHash = u.LastReviewChangedFilesHash
	}
	if u.StopReason != "" {
		r.stopReason = u.StopReason
	}
	r.updates = append(r.updates, u)
	return nil
}

func TestRunImplementationFindingsAutoRecoverUntilExhausted(t *testing.T) {
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
	err := exec.RunImplementation(context.Background(), testPRD)
	if err == nil {
		t.Fatal("RunImplementation() error = nil, want recovery exhausted")
	}
	if !strings.Contains(err.Error(), runstate.StopReasonRecoveryExhausted) {
		t.Fatalf("RunImplementation() error = %v, want %s", err, runstate.StopReasonRecoveryExhausted)
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
	if storyCalls < 2 {
		t.Errorf("story+recovery runner calls = %d, want at least 2", storyCalls)
	}
	if reviewCalls < 1 {
		t.Errorf("review runner calls = %d, want at least 1", reviewCalls)
	}
}

func TestRunImplementationFindingsAutoRecoverAndContinue(t *testing.T) {
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
	reviewCalls := 0
	mock.runFunc = func(_ context.Context, p string, outputCh chan<- runner.OutputLine) error {
		switch {
		case strings.Contains(p, "critical diff review"):
			reviewCalls++
			if reviewCalls == 1 {
				outputCh <- runner.OutputLine{Text: findingsTranscript}
				return nil
			}
			outputCh <- runner.OutputLine{Text: cleanReviewTranscript}
		case prompt.IsRecoveryPrompt(p):
			return nil
		}
		return nil
	}

	exec := NewExecutorWithRunner(cfg, ch, mock)
	if err := exec.RunImplementation(context.Background(), testPRD); err != nil {
		t.Fatalf("RunImplementation() error = %v", err)
	}

	evts := drainEvents(ch)
	var foundSecondStory bool
	for _, e := range evts {
		if ev, ok := e.(EventStoryStarted); ok && ev.Story.ID == "story-2" {
			foundSecondStory = true
		}
	}
	if !foundSecondStory {
		t.Fatal("expected story-2 to start after automatic recovery cleared review findings")
	}
}
